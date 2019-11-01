package csgen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gamewheels/cfgwheel/cfgdef"
)

var namespace = "GameConfig"

// CSGen C#胶水代码生成器
type CSGen struct {
	cfgMap *cfgdef.CfgMap
}

// NewCSGen 构建C#胶水代码生成器
func NewCSGen(cfgMap *cfgdef.CfgMap) *CSGen {
	return &CSGen{
		cfgMap: cfgMap,
	}
}

func genSummary(desc string, tab string) string {
	return tab + "/// <summary>" +
		tab + "/// " + desc +
		tab + "/// </summary>"
}

func genStructName(name string) string {
	if strings.HasSuffix(name, "Table") {
		return name[:len(name)-5] + "Struct"
	} else if strings.HasSuffix(name, "Settings") {
		return name + "Struct"
	}
	return name
}

func getTypeName(field *cfgdef.FieldDef) string {
	if field.IsEnum || field.IsStruct {
		return field.Type
	}
	switch field.Type {
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "int8":
		return "sbyte"
	case "int16":
		return "short"
	case "int32":
		return "int"
	case "int64":
		return "long"
	case "uint8":
		return "byte"
	case "uint16":
		return "ushort"
	case "uint32":
		return "uint"
	case "uint64":
		return "ulong"
	}
	return field.Type
}

// GenType 生成类型名称
func genType(typeName string, isArray bool) string {
	if isArray {
		return typeName + "[]"
	}
	return typeName
}

// GenFileName 生成文件名
func (gen *CSGen) GenFileName(name string) string {
	return name + ".cs"
}

// GenEnum 生成枚举
func (gen *CSGen) GenEnum(name string) string {
	enumDef := gen.cfgMap.EnumMap[name]
	if enumDef == nil || len(enumDef.Items) == 0 {
		fmt.Println("error: ", name, "定义无效")
		return ""
	}
	var buff bytes.Buffer
	buff.WriteString("// Code generated by game config export tool. DO NOT EDIT.")
	buff.WriteString("\r\nnamespace " + namespace)
	buff.WriteString("\r\n{")
	buff.WriteString(genSummary(enumDef.Desc, "\r\n\t"))
	buff.WriteString("\r\n\tpublic enum " + name)
	buff.WriteString("\r\n\t{")
	name2 := name[:len(name)-4]
	for i := 0; i < len(enumDef.Items); i++ {
		item := enumDef.Items[i]
		buff.WriteString(genSummary(item.Desc, "\r\n\t\t"))
		buff.WriteString("\r\n\t\t" + name2 + item.Name + " = " + item.Value + ",")
	}
	buff.WriteString("\r\n\t}")
	buff.WriteString("\r\n}\r\n")
	return buff.String()
}

// GenTable 生成表
func (gen *CSGen) GenTable(name string) string {
	tableDef := gen.cfgMap.TableMap[name]
	if tableDef == nil || len(tableDef.Fields) == 0 {
		fmt.Println("error: ", name, "定义无效")
		return ""
	}

	structName := genStructName(name)
	isTable := strings.HasSuffix(name, "Table")
	isSettings := strings.HasSuffix(name, "Settings")
	var keyField *cfgdef.FieldDef
	var buff bytes.Buffer
	var buff2 bytes.Buffer

	buff.WriteString("// Code generated by game config export tool. DO NOT EDIT.")
	buff.WriteString("\r\nusing System.Runtime.Serialization;")
	buff.WriteString("\r\n\r\nnamespace " + namespace)
	buff.WriteString("\r\n{")

	buff.WriteString(genSummary(tableDef.Desc, "\r\n\t"))
	buff.WriteString("\r\n\t[DataContract]")
	if isTable {
		keyField = tableDef.Fields[tableDef.Key]
		buff.WriteString("\r\n\tpublic class " + structName + " : IConfigStruct<" + getTypeName(keyField) + ">")
	} else {
		buff.WriteString("\r\n\tpublic class " + structName)
	}
	buff.WriteString("\r\n\t{")

	for i := 0; i < len(tableDef.Fields); i++ {
		field := tableDef.Fields[i]
		if field.Name != "" && field.Type != "" &&
			(field.IsKey || field.UseFor == "A" || field.UseFor == cfgdef.ExportFlags.UseFor) {
			typeName := getTypeName(field)
			buff.WriteString(genSummary(field.Desc, "\r\n\t\t"))
			buff.WriteString("\r\n\t\t[DataMember]")
			buff.WriteString("\r\n\t\tpublic " + genType(typeName, field.IsArray) + " " + field.Name + " { get; private set; }")
			if field.FTable != "" {
				relateName := field.Name + "2" + field.FTable
				buff.WriteString(genSummary(field.Name+" --> "+field.FTable, "\r\n\t\t"))
				buff.WriteString("\r\n\t\tpublic " + genType(field.FTable+"Struct", field.IsArray) + " " + relateName + " { get; private set; }")
				if field.IsArray {
					buff2.WriteString("\r\n\t\t\t" + relateName + " = new " + field.FTable + "Struct[" + field.Name + ".Length];")
					buff2.WriteString("\r\n\t\t\tfor (int i = 0; i < " + field.Name + ".Length; ++i)")
					buff2.WriteString("\r\n\t\t\t{")
					buff2.WriteString("\r\n\t\t\t\t" + relateName + "[i] = Facade." + field.FTable + "Table[" + field.Name + "[i]];")
					buff2.WriteString("\r\n\t\t\t}")
				} else {
					buff2.WriteString("\r\n\t\t\t" + relateName + " = Facade." + field.FTable + "Table[" + field.Name + "];")
				}
			}
		}
	}
	if isTable {
		buff.WriteString("\r\n\r\n\t\tpublic " + getTypeName(keyField) + " GetKey() { return " + keyField.Name + "; }")
	}
	buff.WriteString("\r\n\r\n\t\tpublic void Relate()")
	buff.WriteString("\r\n\t\t{")
	buff.WriteString(buff2.String())
	buff.WriteString("\r\n\t\t}")
	buff.WriteString("\r\n\t}")

	if isTable {
		buff.WriteString("\r\n\r\n\tpublic partial class Facade")
		buff.WriteString("\r\n\t{")
		buff.WriteString(genSummary(tableDef.Desc, "\r\n\t\t"))
		buff.WriteString("\r\n\t\tpublic static DataTable<" + getTypeName(keyField) + ", " + structName + "> " +
			name + " = DataTable<" + getTypeName(keyField) + ", " + structName + ">.Instance;")
		buff.WriteString("\r\n\t}")
	} else if isSettings {
		buff.WriteString("\r\n\r\n\tpublic partial class Facade")
		buff.WriteString("\r\n\t{")
		buff.WriteString(genSummary(tableDef.Desc, "\r\n\t\t"))
		buff.WriteString("\r\n\t\tpublic static " + structName + " " + name + ";")
		buff.WriteString("\r\n\t}")
	}

	buff.WriteString("\r\n}\r\n")
	return buff.String()
}
