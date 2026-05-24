package clause

// Pagination 表示分页信息，包含 LIMIT 和 OFFSET。
// Limit 为指针类型，nil 表示不限制；Offset 默认为 0。
type Pagination struct {
	Limit  *int
	Offset int
}

func (p Pagination) Build(builder Builder) {
	// 构建 LIMIT 部分
	if p.Limit != nil && *p.Limit > 0 {
		builder.WriteString(" LIMIT ")
		builder.AddVar(builder, *p.Limit)
	}

	// 构建 OFFSET 部分
	if p.Offset > 0 {
		builder.WriteString(" OFFSET ")
		builder.AddVar(builder, p.Offset)
	}
}
