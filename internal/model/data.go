package model

// DataSheet 表示一个数据表
type DataSheet struct {
	Name    string                   // 表名
	Columns []ColumnInfo             // 列信息
	Rows    []map[string]interface{} // 行数据
	Meta    map[string]interface{}   // 元数据
}

// ColumnInfo 表示列信息
type ColumnInfo struct {
	Name     string      // 列名
	Type     string      // 数据类型
	Comment  string      // 注释
	Required bool        // 是否必填
	Default  interface{} // 默认值
	Options  []string    // 可选值（枚举）
	Ref      *RefInfo    // 引用信息
}

// RefInfo 表示引用关系
type RefInfo struct {
	Sheet  string // 引用的表名
	Column string // 引用的列名
}

// ConvertResult 表示转换结果
type ConvertResult struct {
	FileName string // 输出文件名
	Content  []byte // 转换后的内容
	Format   string // 格式类型
}

// ErrorInfo 表示错误信息
type ErrorInfo struct {
	Sheet  string // 表名
	Row    int    // 行号
	Column string // 列名
	Msg    string // 错误消息
}
