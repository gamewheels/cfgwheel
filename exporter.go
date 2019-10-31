package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/gamewheels/cfgwheel/cfgdef"
	"github.com/tealeg/xlsx"
)

// xlsMap Excel配置文件列表
var xlsMap = make(map[int]string)
var cfgMap = cfgdef.NewCfgMap()

// loadXlsList 加载Excel配置文件列表
func loadXlsList(pathname string) {
	fileInfo, err := os.Stat(pathname)
	if err != nil {
		return
	}

	if fileInfo.IsDir() {
		all, _ := ioutil.ReadDir(pathname)
		for _, f := range all {
			fn := f.Name()
			ext := strings.ToLower(path.Ext(fn))
			if f.IsDir() {
				loadXlsList(pathname + "/" + fn)
			} else if !strings.HasPrefix(fn, "~$") &&
				(ext == ".xls" || ext == ".xlsx") {
				xlsMap[len(xlsMap)] = pathname + "/" + fn
			}
		}
	} else {
		xlsMap[len(xlsMap)] = pathname
	}
}

func lineTrim(s string) string {
	return cfgdef.Trim(strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\n", " "))
}

// loadEnumCfg 加载枚举配置
func loadEnumCfg(sheet *xlsx.Sheet) {
	name := sheet.Name
	if sheet.MaxCol < 3 || sheet.MaxRow < 3 {
		fmt.Println("error: enum", name, "格式不正确")
		return
	}
	if _, ok := cfgMap.EnumMap[name]; ok {
		fmt.Println("error: enum", name, "重复定义")
		return
	}
	enumDef := cfgdef.NewEnumDef(name)
	enumDef.Desc = lineTrim(sheet.Rows[0].Cells[0].String())
	for i := 2; i < sheet.MaxRow; i++ {
		cells := sheet.Rows[i].Cells
		if len(cells) > 2 && cells[0].String() != "" && cells[1].String() != "" {
			item := &cfgdef.EnumItem{
				Name:  cfgdef.Trim(cells[0].String()),
				Value: cfgdef.Trim(cells[1].String()),
				Desc:  lineTrim(cells[2].String()),
			}
			enumDef.Items[len(enumDef.Items)] = item
			enumDef.ItemsMap[item.Name] = item
		}
	}
	cfgMap.EnumMap[name] = enumDef
}

// loadTableCfg 加载表格配置
func loadTableCfg(sheet *xlsx.Sheet) {
	name := sheet.Name
	isTable := strings.HasSuffix(name, "Table")
	if sheet.MaxCol < 2 || sheet.MaxRow < 5 {
		fmt.Println("error:", name, "格式不正确")
		return
	}
	if _, ok := cfgMap.TableMap[name]; ok {
		fmt.Println("error:", name, "重复定义")
		return
	}

	//解析表结构
	tableDef := cfgdef.NewTableDef(name)
	tableDef.Desc = lineTrim(sheet.Rows[0].Cells[0].String())
	for i := 0; i < sheet.MaxCol; i++ {
		fullType := cfgdef.GetFullFieldType(sheet.Rows[3].Cells[i].String())
		if fullType == "?" {
			fmt.Println("error:", name, "字段类型无效", sheet.Rows[3].Cells[i].String())
			fullType = ""
		}
		constraint := sheet.Rows[2].Cells[i].String() // 字段约束
		field := &cfgdef.FieldDef{
			Name:     cfgdef.Trim(sheet.Rows[4].Cells[i].String()),
			Type:     cfgdef.GetFieldType(fullType),
			Desc:     lineTrim(sheet.Rows[1].Cells[i].String()),
			IsArray:  strings.HasPrefix(fullType, "[]"),
			IsStruct: strings.HasSuffix(fullType, "Struct"),
			IsEnum:   strings.HasSuffix(fullType, "Enum"),
		}
		//解析字段约束
		temp1 := strings.Split(constraint, ";")
		for j := 0; j < len(temp1); j++ {
			Cmd := temp1[j]
			//主键
			if Cmd == "K" && tableDef.Key < 0 && field.Type != "" {
				field.IsKey = true
				tableDef.Key = i
				if field.IsArray {
					fmt.Println("error:", name, "主键字段不可为数组")
				}
			}
			//字段用途 A:前后端通用 S:后端 C:前端
			if Cmd == "A" || Cmd == "S" || Cmd == "C" {
				field.UseFor = Cmd
			}
			//字符串或者数组长度范围
			if strings.HasPrefix(Cmd, "L[") && strings.HasSuffix(Cmd, "]") {
				err := json.Unmarshal([]byte(Cmd[1:]), &field.Len)
				if err != nil {
					fmt.Println("error:", name, "字段约束定义有误", Cmd)
				}
			}
			//取值范围
			if strings.HasPrefix(Cmd, "R[") && strings.HasSuffix(Cmd, "]") {
				err := json.Unmarshal([]byte(Cmd[1:]), &field.Range)
				if err != nil {
					fmt.Println("error:", name, "字段约束定义有误", Cmd)
				}
			}
			if strings.HasPrefix(Cmd, "F[") && strings.HasSuffix(Cmd, "]") {
				field.FTable = Cmd[2 : len(Cmd)-1]
			}
		}
		tableDef.Fields[i] = field
		tableDef.FieldsMap[field.Name] = field
	}

	if isTable && tableDef.Key < 0 {
		fmt.Println("error:", name, "缺少主键")
		return
	}

	//加载数据
	fields := len(tableDef.Fields)
	for i := 5; i < sheet.MaxRow; i++ {
		if strings.HasSuffix(name, "Struct") {
			break
		}
		cells := sheet.Rows[i].Cells
		data := make([]string, fields)
		for j := 0; j < fields; j++ {
			if j < len(cells) {
				data[j] = cells[j].String()
			}
		}
		tableDef.Data[len(tableDef.Data)] = data
		if strings.HasSuffix(name, "Settings") {
			break
		}
		key := cells[tableDef.Key].String()
		tableDef.DataMap[key] = data
	}

	cfgMap.TableMap[name] = tableDef
}

// loadAllCfg 加载全部配置
func loadAllCfg(filepath string) {
	fmt.Println("加载配置文件:", filepath, "...")
	xls, err := xlsx.OpenFile(filepath)
	if err != nil {
		fmt.Println("Failed to open ", filepath)
		return
	}
	for _, sheet := range xls.Sheets {
		fmt.Println("加载:", sheet.Name, "...")
		switch {
		case strings.HasSuffix(sheet.Name, "Enum"):
			loadEnumCfg(sheet)
		case strings.HasSuffix(sheet.Name, "Settings"),
			strings.HasSuffix(sheet.Name, "Struct"),
			strings.HasSuffix(sheet.Name, "Table"):
			loadTableCfg(sheet)
		}
	}
}

func saveToFile(filename string, s string) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC, 0600)
	if err == nil {
		f.Write([]byte(s))
	} else {
		fmt.Println(err)
	}
	f.Close()
}

// genCode 生成胶水代码或者配置数据
func genCode(gen cfgdef.Generator) {
	for n := range cfgMap.EnumMap {
		filename := gen.GenFileName(n)
		if filename != "" {
			fmt.Println("生成:", n, "...")
			saveToFile(cfgdef.ExportFlags.OutputPath+"/"+filename, gen.GenEnum(n))
		}
	}
	for n := range cfgMap.TableMap {
		filename := gen.GenFileName(n)
		if filename != "" {
			fmt.Println("生成:", n, "...")
			saveToFile(cfgdef.ExportFlags.OutputPath+"/"+filename, gen.GenTable(n))
		}
	}
}
