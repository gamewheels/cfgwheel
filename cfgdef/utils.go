package cfgdef

import "strings"

// Trim 去掉字符串首尾的空白
func Trim(s string) string {
	return strings.Trim(s, " \t\n\r")
}

// IsJSONArray 初略判断是否是JSON数组
func IsJSONArray(s string) bool {
	return strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]")
}

// IsJSONObject 初略判断是否是JSON对象
func IsJSONObject(s string) bool {
	return strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")
}

// GetFieldType 获得字段类型
func GetFieldType(typeName string) string {
	if strings.HasPrefix(typeName, "[]") {
		return typeName[2:]
	}
	return typeName
}

// GetArraySymbol 获得数组标记
func GetArraySymbol(isArray bool) string {
	if isArray {
		return "[]"
	}
	return ""
}

// GetFullFieldType 获得包含数组标记的字段类型
func GetFullFieldType(typeName string) string {
	typeName = Trim(typeName)
	arr := ""
	fieldType := strings.ToLower(typeName)
	if strings.HasPrefix(typeName, "[]") {
		typeName = Trim(typeName[2:])
		fieldType = strings.ToLower(typeName)
		arr = "[]"
	}
	switch fieldType {
	case "string", "int8":
		return arr + fieldType
	case "bool", "boolean":
		return arr + "bool"
	case "byte", "uint8":
		return arr + "uint8"
	case "short", "int16":
		return arr + "int16"
	case "ushort", "uint16":
		return arr + "uint16"
	case "int", "int32":
		return arr + "int32"
	case "uint", "uint32":
		return arr + "uint32"
	case "long", "int64":
		return arr + "int64"
	case "ulong", "uint64":
		return arr + "uint64"
	case "float", "float32":
		return arr + "float32"
	case "double", "number", "float64":
		return arr + "float64"
	case "":
		return typeName
	}
	if (strings.HasSuffix(typeName, "Enum") && len(fieldType) > 4) ||
		(strings.HasSuffix(typeName, "Struct") && len(fieldType) > 6) {
		return arr + typeName
	}
	return "?"
}
