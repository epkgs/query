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

// OrderBys 是排序条件列表，表示一组 ORDER BY 子句。
// 它是 []*OrderBy 的类型别名，提供了 Map 和 MapColumn 方法用于遍历和转换排序条件。
type OrderBys []*OrderBy

// Build 构建 ORDER BY 子句。
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

// Map 遍历排序条件列表，并生成新的排序条件列表
//
// mapper 为排序条件遍历函数，返回nil表示移除该排序条件
func (obs OrderBys) Map(mapper func(o OrderBy) *OrderBy) OrderBys {
	result := make(OrderBys, 0, len(obs))

	for _, ob := range obs {

		if ob == nil {
			continue
		}

		newOb := mapper(*ob)
		if newOb == nil {
			continue
		}
		result = append(result, newOb)
	}

	return result
}
