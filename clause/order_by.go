package clause

// OrderBy 表示排序条件
type OrderBy struct {
	Column string
	Desc   bool
}

func (o OrderBy) Build(builder Builder) {
	builder.WriteQuoted(o.Column)
	if o.Desc {
		builder.WriteString(" DESC")
	} else {
		builder.WriteString(" ASC")
	}
}

type OrderBys []OrderBy

// Build build where clause
func (o OrderBys) Build(builder Builder) {

	if len(o) > 0 {
		builder.WriteString(" ORDER BY ")
	}

	for idx, order := range o {
		if idx > 0 {
			builder.WriteString(", ")
		}

		order.Build(builder)
	}
}
