package clause

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
