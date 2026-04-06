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

type OrderBys []*OrderBy

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

func (obs OrderBys) Map(mapper func(*OrderBy) *OrderBy) OrderBys {

	result := make(OrderBys, 0, len(obs))

	for _, ob := range obs {
		newOb := mapper(&OrderBy{
			Column: ob.Column,
			Desc:   ob.Desc,
		})
		if newOb == nil {
			continue
		}
		result = append(result, newOb)
	}

	return result
}

func (obs OrderBys) MapColumn(mapper func(string, bool) (string, bool)) OrderBys {

	result := make(OrderBys, 0, len(obs))

	for _, ob := range obs {
		column, desc := mapper(ob.Column, ob.Desc)
		if column == "" {
			continue
		}
		result = append(result, &OrderBy{
			Column: column,
			Desc:   desc,
		})
	}

	return result
}
