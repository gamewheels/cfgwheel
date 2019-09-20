package jsongen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gamewheels/cfgwheel/cfgdef"
)

// JSONGen json生成器
type JSONGen struct {
	cfgMap *cfgdef.CfgMap
}

var currentLine = 0

// NewJSONGen 构建json生成器
func NewJSONGen(cfgMap *cfgdef.CfgMap) *JSONGen {
	return &JSONGen{
		cfgMap: cfgMap,
	}
}

func genStructName(name string) string {
	if strings.HasSuffix(name, "Table") {
		return name[:len(name)-5] + "Struct"
	} else if strings.HasSuffix(name, "Settings") {
		return name + "Struct"
	}
	return name
}

// GenFileName 生成文件名
func (gen *JSONGen) GenFileName(name string) string {
	if strings.HasSuffix(name, "Enum") || strings.HasSuffix(name, "Struct") {
		return ""
	}
	return name + ".json"
}

// GenEnum 生成枚举
func (gen *JSONGen) GenEnum(name string) string {
	return ""
}

// GenTable 生成表
func (gen *JSONGen) GenTable(name string) string {
	tableDef := gen.cfgMap.TableMap[name]
	if tableDef == nil || len(tableDef.Fields) == 0 {
		fmt.Println("error: ", name, "定义无效")
		return ""
	}

	isSettings := strings.HasSuffix(name, "Settings")

	if isSettings {
		if len(tableDef.Data) < 1 {
			fmt.Println("error: ", name, "缺少配置数据")
			return ""
		}
		return gen.genStructValue(tableDef.Data[0], tableDef)
	}

	var buff bytes.Buffer
	buff.WriteString("[")
	sp := ""
	for i := 0; i < len(tableDef.Data); i++ {
		currentLine = i + 6
		buff.WriteString(sp)
		buff.WriteString(gen.genStructValue(tableDef.Data[i], tableDef))
		sp = ",\n"
	}
	buff.WriteString("]")
	return buff.String()
}

// 生成枚举值
func (gen *JSONGen) genEnumValue(s string, field *cfgdef.FieldDef) string {
	enumDef, ok := gen.cfgMap.EnumMap[field.Type]
	if ok {
		s = cfgdef.Trim(s)
		value, ok := enumDef.ItemsMap[s]
		if ok {
			return value.Value
		}
	}
	fmt.Printf("error: line[%d] 枚举%s.%s未定义\n", currentLine, field.Type, s)
	return toIntValue(s)
}

// 转换为true/false
func toBoolValue(s string) string {
	s = strings.ToLower(cfgdef.Trim(s))
	if s == "true" {
		return "true"
	}
	value, err := strconv.Atoi(s)
	if err == nil && value != 0 {
		return "true"
	}
	return "false"
}

//转换字符串字段值
func toStringValue(s string) string {
	value, _ := json.Marshal(s)
	return string(value)
}

//转换为数字字段值
func toNumberValue(s string) string {
	var value float64
	s = cfgdef.Trim(s)
	err := json.Unmarshal([]byte(s), &value)
	if err == nil {
		return s
	}
	fmt.Printf("error: line[%d] %s 转换为数字失败\n", currentLine, s)
	return s
}

//转换为整形字段值
func toIntValue(s string) string {
	var value int64
	s = cfgdef.Trim(s)
	err := json.Unmarshal([]byte(s), &value)
	if err == nil {
		return s
	}
	fmt.Printf("error: line[%d] %s 转换为整数失败\n", currentLine, s)
	return s
}

//转换为整形字段值
func toUIntValue(s string) string {
	var value uint64
	s = cfgdef.Trim(s)
	err := json.Unmarshal([]byte(s), &value)
	if err == nil {
		return s
	}
	fmt.Printf("error: line[%d] %s 转换为正整数失败\n", currentLine, s)
	return s
}

//生成字段值
func (gen *JSONGen) genFieldValue(s string, field *cfgdef.FieldDef) string {
	if field.IsStruct {
		return gen.genStructFromString(s, field.Type)
	} else if field.IsEnum {
		return gen.genEnumValue(s, field)
	}
	switch field.Type {
	case "bool":
		return toBoolValue(s)
	case "string":
		return toStringValue(s)
	case "float32", "float64":
		return toNumberValue(s)
	case "uint8", "uint16", "uint32", "uint64":
		return toUIntValue(s)
	}
	return toIntValue(s)
}

//从JSON生成字段值
func (gen *JSONGen) genFieldValueFromJSON(s string, field *cfgdef.FieldDef) string {
	if field.IsStruct {
		return gen.genStructFromString(s, field.Type)
	} else if field.IsEnum {
		return toIntValue(s)
	}

	switch field.Type {
	case "bool":
		return toBoolValue(s)
	case "float32", "float64":
		return toNumberValue(s)
	case "int8", "int16", "int32", "int64":
		return toIntValue(s)
	case "uint8", "uint16", "uint32", "uint64":
		return toUIntValue(s)
	}

	var temp2 string
	err := json.Unmarshal([]byte(s), &temp2)
	if err == nil {
		return toStringValue(temp2)
	}
	return "\"" + s + "\""
}

//生成数组
func (gen *JSONGen) genArrayValue(s string, field *cfgdef.FieldDef) string {
	s = cfgdef.Trim(s)
	if s == "" {
		return "null"
	}
	var temp []cfgdef.AnyField
	if json.Unmarshal([]byte(s), &temp) != nil {
		fmt.Printf("error: line[%d] %s 转换为数组失败\n", currentLine, s)
		return "null"
	}

	var buff bytes.Buffer
	buff.WriteString("[")
	sp := ""
	for _, v := range temp {
		buff.WriteString(sp + gen.genFieldValueFromJSON(v.Value, field))
		sp = ","
	}
	buff.WriteString("]")
	return buff.String()
}

//从字符串生成结构体
func (gen *JSONGen) genStructFromString(s string, typeName string) string {
	structDef, ok := gen.cfgMap.TableMap[typeName]
	if !ok {
		fmt.Println("error: ", typeName, " 未定义")
		return "null"
	}
	s = cfgdef.Trim(s)
	if cfgdef.IsJSONArray(s) {
		var temp []cfgdef.AnyField
		if json.Unmarshal([]byte(s), &temp) != nil {
			fmt.Printf("error: line[%d] %s 转换为 %s 失败\n", currentLine, s, typeName)
			return "null"
		}
		var buff bytes.Buffer
		buff.WriteString("{")
		sp := ""
		for i := 0; i < len(temp) && i < len(structDef.Fields); i++ {
			field := structDef.Fields[i]
			buff.WriteString(sp + "\"" + field.Name + "\":" + gen.genFieldValueFromJSON(temp[i].Value, field))
			sp = ","
		}
		buff.WriteString("}")
		return buff.String()
	} else if cfgdef.IsJSONObject(s) {
		var temp map[string]cfgdef.AnyField
		if json.Unmarshal([]byte(s), &temp) != nil {
			fmt.Printf("error: line[%d] %s 转换为 %s 失败\n", currentLine, s, typeName)
			return "null"
		}
		var buff bytes.Buffer
		buff.WriteString("{")
		sp := ""
		for n, v := range temp {
			field, ok := structDef.FieldsMap[n]
			if ok {
				buff.WriteString(sp + "\"" + n + "\":" + gen.genFieldValueFromJSON(v.Value, field))
				sp = ","
			}
		}
		buff.WriteString("}")
		return buff.String()
	}
	return "null"
}

// genStructValue 生成结构体
func (gen *JSONGen) genStructValue(cols []string, structDef *cfgdef.TableDef) string {
	var buff bytes.Buffer
	sp := ""
	buff.WriteString("{")
	for j := 0; j < len(cols); j++ {
		field := structDef.Fields[j]
		if field.Name != "" && field.Type != "" &&
			(field.IsKey || field.UseFor == "A" || field.UseFor == cfgdef.ExportFlags.UseFor) {
			var value string
			if field.IsArray {
				value = gen.genArrayValue(cols[j], field)
				gen.checkArray(value, field)
			} else {
				value = gen.genFieldValue(cols[j], field)
				gen.checkValue(value, field)
			}
			buff.WriteString(sp + "\"" + field.Name + "\":" + value)
			sp = ","
		}
	}
	buff.WriteString("}")
	return buff.String()
}

//外键关联检查
func (gen *JSONGen) checkFTable(s string, field *cfgdef.FieldDef) {
	if ft, ok := gen.cfgMap.TableMap[field.FTable+"Table"]; ok {
		if _, ok := ft.DataMap[s]; !ok {
			fmt.Println("error: 没找到", field.FTable, s)
		}
	} else {
		fmt.Println("error: 缺少外键关联表", field.FTable)
	}
}

//取值范围检查
func (gen *JSONGen) checkRange(s string, field *cfgdef.FieldDef) {
	var v float64
	if json.Unmarshal([]byte(s), &v) == nil {
		ok := true
		if len(field.Range) == 1 {
			if v > field.Range[0] {
				ok = false
			}
		} else if len(field.Range) > 1 {
			if v < field.Range[0] || v > field.Range[1] {
				ok = false
			}
		}
		if !ok {
			fmt.Println("error: 字段取值范围错误", field.Name, field.Range, s)
		}
	} else {
		fmt.Println("error: 字段值填写错误", field.Name, s)
	}
}

//长度范围检查
func (gen *JSONGen) checkLen(s string, field *cfgdef.FieldDef) {
	if field.IsArray {
		var varr []cfgdef.AnyField
		if err := json.Unmarshal([]byte(s), &varr); err != nil {
			fmt.Println("error:", err)
		} else {
			ok := true
			l := uint(len(varr))
			if len(field.Len) == 1 {
				if l > field.Len[0] {
					ok = false
				}
			} else if len(field.Len) > 1 {
				if l < field.Len[0] || l > field.Len[1] {
					ok = false
				}
			}
			if !ok {
				fmt.Println("error: 字符串长度范围错误", field.Name, field.Len, s, l)
			}
		}
	} else if field.Type == "string" {
		ok := true
		v := ""
		json.Unmarshal([]byte(s), &v)
		l := uint(len(v))
		if len(field.Len) == 1 {
			if l > field.Len[0] {
				ok = false
			}
		} else if len(field.Len) > 1 {
			if l < field.Len[0] || l > field.Len[1] {
				ok = false
			}
		}
		if !ok {
			fmt.Println("error: 字符串长度范围错误", field.Name, field.Len, v, l)
		}
	}
}

//检查数组字段值
func (gen *JSONGen) checkArray(s string, field *cfgdef.FieldDef) {
	if field.Len != nil {
		gen.checkLen(s, field)
	}
	var va []cfgdef.AnyField
	if err := json.Unmarshal([]byte(s), &va); err != nil {
		fmt.Println("error:", err)
	} else {
		for _, v := range va {
			if field.FTable != "" {
				gen.checkFTable(v.Value, field)
			}
			if field.Range != nil {
				gen.checkRange(v.Value, field)
			}
		}
	}
}

//检查字段值
func (gen *JSONGen) checkValue(s string, field *cfgdef.FieldDef) {
	if field.FTable != "" {
		gen.checkFTable(s, field)
	}
	if field.Range != nil {
		gen.checkRange(s, field)
	}
	if field.Len != nil {
		gen.checkLen(s, field)
	}
}
