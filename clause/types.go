package clause

// Valuer 表示一个可以延迟求值的值。
// 当 WHERE 条件中的值为 nil 或实现了 Valuer 接口时，
// 系统会调用 Value() 方法来获取实际值，
// 用于判断是否应该生成 IS NULL 或 IS NOT NULL 表达式。
type Valuer interface {
	Value() (any, error)
}
