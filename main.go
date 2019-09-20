package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gamewheels/cfgwheel/cfgdef"
	"github.com/gamewheels/cfgwheel/cppgen"
	"github.com/gamewheels/cfgwheel/csgen"
	"github.com/gamewheels/cfgwheel/gogen"
	"github.com/gamewheels/cfgwheel/jsongen"
)

func repairPath(p *string, create bool) {
	*p = strings.ReplaceAll(*p, "\\", "/")
	if !strings.HasSuffix(*p, "/") {
		*p = *p + "/"
	}
	if create {
		os.MkdirAll(*p, os.ModePerm)
	}
}

//根据自己的清空修改吧
func main() {
	flag.StringVar(&cfgdef.ExportFlags.XLSPath, "xls", "./excel/", "Excel配置源路径")
	flag.StringVar(&cfgdef.ExportFlags.JSONPath, "json", "./json/", "JSON输出路径")
	flag.StringVar(&cfgdef.ExportFlags.GoPath, "go", "./src/gameconfig/", "GO胶水代码输出路径")
	flag.StringVar(&cfgdef.ExportFlags.CPPPath, "cpp", "./gameconfig/", "CPP胶水代码输出路径")
	flag.StringVar(&cfgdef.ExportFlags.CSPath, "cs", "./gameconfig/", "C#胶水代码输出路径")
	flag.StringVar(&cfgdef.ExportFlags.UseFor, "use", "S", "S:服务端使用 C:客户端使用")
	flag.Parse()

	repairPath(&cfgdef.ExportFlags.XLSPath, false)
	loadXlsList(cfgdef.ExportFlags.XLSPath)
	for _, fn := range xlsMap {
		loadAllCfg(fn)
	}

	fmt.Println("\n生成Golang胶水代码 ...")
	repairPath(&cfgdef.ExportFlags.GoPath, true)
	cfgdef.ExportFlags.OutputPath = cfgdef.ExportFlags.GoPath
	genCode(gogen.NewGoGen(cfgMap))

	fmt.Println("\n生成C++胶水代码 ...")
	repairPath(&cfgdef.ExportFlags.CPPPath, true)
	cfgdef.ExportFlags.OutputPath = cfgdef.ExportFlags.CPPPath
	genCode(cppgen.NewCPPGen(cfgMap))

	fmt.Println("\n生成CS胶水代码 ...")
	repairPath(&cfgdef.ExportFlags.CSPath, true)
	cfgdef.ExportFlags.OutputPath = cfgdef.ExportFlags.CSPath
	genCode(csgen.NewCSGen(cfgMap))

	fmt.Println("\n生成JSON数据 ...")
	repairPath(&cfgdef.ExportFlags.JSONPath, true)
	cfgdef.ExportFlags.OutputPath = cfgdef.ExportFlags.JSONPath
	genCode(jsongen.NewJSONGen(cfgMap))
}
