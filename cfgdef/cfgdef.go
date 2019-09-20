package cfgdef

// ExportFlags 导出参数
var ExportFlags = struct {
	XLSPath    string
	OutputPath string
	JSONPath   string
	GoPath     string
	CPPPath    string
	CSPath     string
	UseFor     string
}{}

// EnumItem 枚举项
type EnumItem struct {
	Name  string
	Value string
	Desc  string
}

// EnumDef 枚举信息
type EnumDef struct {
	Name     string               // 名称
	Desc     string               // 描述
	Items    map[int]*EnumItem    // 枚举项
	ItemsMap map[string]*EnumItem // 枚举项
}

// FieldDef 字段
type FieldDef struct {
	Name     string    // 字段名
	Type     string    // 字段类型
	Desc     string    // 字段说明
	IsArray  bool      // 是否是数组
	IsKey    bool      // 是否是键值
	IsEnum   bool      // 是否是枚举
	IsStruct bool      // 是否是结构体
	UseFor   string    // 字段用途
	Len      []uint    // 数组元素个数或字符串长度范围
	Range    []float64 // 数值取值范围
	FTable   string    // 外键关联表
}

// TableDef 表格定义
type TableDef struct {
	Name      string               // 名称
	Desc      string               // 描述
	Key       int                  // 主键字段
	Fields    map[int]*FieldDef    // 字段
	FieldsMap map[string]*FieldDef // 字段
	Data      map[int][]string     // 数据
	DataMap   map[string][]string  // 数据
}

// CfgMap 配置信息
type CfgMap struct {
	EnumMap  map[string]*EnumDef
	TableMap map[string]*TableDef
}

// NewEnumDef 构建EnumDef
func NewEnumDef(name string) *EnumDef {
	return &EnumDef{
		Name:     name,
		Items:    make(map[int]*EnumItem),
		ItemsMap: make(map[string]*EnumItem),
	}
}

// NewTableDef 构建NewTableDef
func NewTableDef(name string) *TableDef {
	return &TableDef{
		Name:      name,
		Key:       -1,
		Fields:    make(map[int]*FieldDef),
		FieldsMap: make(map[string]*FieldDef),
		Data:      make(map[int][]string),
		DataMap:   make(map[string][]string),
	}
}

// NewCfgMap 构建CfgMap
func NewCfgMap() *CfgMap {
	return &CfgMap{
		EnumMap:  make(map[string]*EnumDef),
		TableMap: make(map[string]*TableDef),
	}
}

// AnyField AnyField
type AnyField struct {
	Value string
}

// MarshalJSON MarshalJSON
func (field *AnyField) MarshalJSON() ([]byte, error) {
	return []byte(field.Value), nil
}

// UnmarshalJSON UnmarshalJSON
func (field *AnyField) UnmarshalJSON(value []byte) error {
	field.Value = string(value)
	return nil
}
