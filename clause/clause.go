// Package clause 定义了查询构建器的底层抽象组件。
//
// 该包提供了 SQL 子句的抽象表示，包括表达式（Expression）、条件（Where）、
// 排序（OrderBy）和分页（Pagination）。所有组件都实现了 Expression 接口，
// 可通过 Builder 接口写入到不同的目标（标准 SQL、GORM、Ent 等）。
//
// 该包的类型也被适配器包用于将查询组件转换为特定 ORM 的查询条件。
package clause

// Writer 是一个基础写入接口，定义了字节和字符串的写入方法。
type Writer interface {
	WriteByte(byte) error
	WriteString(string) (int, error)
}

// Builder 定义了查询构建器的写入接口。
// 它扩展了 Writer 接口，提供了带引号写入、变量绑定和错误收集的能力。
// 不同的适配器（GORM、Ent、标准 SQL）各自实现此接口，
// 从而将抽象的查询组件转换为目标格式。
type Builder interface {
	Writer
	WriteQuoted(field interface{})
	AddVar(Writer, ...interface{})
	AddError(error) error
}
