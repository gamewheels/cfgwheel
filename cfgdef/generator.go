package cfgdef

// Generator 胶水代码生成器接口
type Generator interface {
	// GenFileName 生成文件名
	GenFileName(name string) string

	// GenEnum 生成枚举
	GenEnum(name string) string

	// GenTable 生成表
	GenTable(name string) string
}
