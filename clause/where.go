package clause

const (
	AndWithSpace = " AND "
	OrWithSpace  = " OR "
)

// Where where clause
type Where struct {
	Exprs []Expression
}

// Build build where clause
func (w Where) Build(builder Builder) {
	exprs := w.Exprs
	if len(exprs) == 1 {
		if andExpr, ok := exprs[0].(AndExpr); ok {
			exprs = andExpr.Exprs
		}
	}
	if len(exprs) > 0 {
		builder.WriteString(" WHERE ")
		buildExprs(exprs, builder, AndWithSpace)
	}
}

func (w *Where) Merge(where Where) *Where {
	if len(where.Exprs) > 0 {
		exprs := make([]Expression, len(w.Exprs)+len(where.Exprs))
		copy(exprs, w.Exprs)
		copy(exprs[len(w.Exprs):], where.Exprs)
		w.Exprs = exprs
	}
	return w
}

func And(exprs ...Expression) Expression {
	if len(exprs) == 0 {
		return nil
	}

	if len(exprs) == 1 {
		if _, ok := exprs[0].(OrExpr); !ok {
			return exprs[0]
		}
	}

	return AndExpr{Exprs: exprs}
}

type AndExpr struct {
	Exprs []Expression
}

func (and AndExpr) Build(builder Builder) {
	if len(and.Exprs) > 1 {
		builder.WriteByte('(')
		buildExprs(and.Exprs, builder, " AND ")
		builder.WriteByte(')')
	} else {
		buildExprs(and.Exprs, builder, " AND ")
	}
}

func Or(exprs ...Expression) Expression {
	if len(exprs) == 0 {
		return nil
	}
	return OrExpr{Exprs: exprs}
}

type OrExpr struct {
	Exprs []Expression
}

func (or OrExpr) Build(builder Builder) {
	if len(or.Exprs) > 1 {
		builder.WriteByte('(')
		buildExprs(or.Exprs, builder, " OR ")
		builder.WriteByte(')')
	} else {
		buildExprs(or.Exprs, builder, " OR ")
	}
}

func Not(exprs ...Expression) Expression {
	if len(exprs) == 0 {
		return nil
	}
	if len(exprs) == 1 {
		if andCondition, ok := exprs[0].(AndExpr); ok {
			exprs = andCondition.Exprs
		}
	}
	return NotExpr{Exprs: exprs}
}

type NotExpr struct {
	Exprs []Expression
}

func (not NotExpr) Build(builder Builder) {
	anyNegationBuilder := false
	for _, c := range not.Exprs {
		if _, ok := c.(NegationExpressionBuilder); ok {
			anyNegationBuilder = true
			break
		}
	}

	if anyNegationBuilder {
		if len(not.Exprs) > 1 {
			builder.WriteByte('(')
		}

		for idx, c := range not.Exprs {
			if idx > 0 {
				builder.WriteString(" AND ")
			}

			if negationBuilder, ok := c.(NegationExpressionBuilder); ok {
				negationBuilder.NegationBuild(builder)
			} else {
				c.Build(builder)
			}
		}

		if len(not.Exprs) > 1 {
			builder.WriteByte(')')
		}
	} else {
		builder.WriteString("NOT ")
		if len(not.Exprs) > 1 {
			builder.WriteByte('(')
		}

		for idx, c := range not.Exprs {
			if idx > 0 {
				switch c.(type) {
				case OrExpr:
					builder.WriteString(" OR ")
				default:
					builder.WriteString(" AND ")
				}
			}

			c.Build(builder)
		}

		if len(not.Exprs) > 1 {
			builder.WriteByte(')')
		}
	}
}

func buildExprs(exprs []Expression, builder Builder, joinCond string) {

	for idx, expr := range exprs {
		if idx > 0 {
			if v, ok := expr.(OrExpr); ok && len(v.Exprs) == 1 {
				builder.WriteString(OrWithSpace)
			} else {
				builder.WriteString(joinCond)
			}
		}

		expr.Build(builder)
	}
}
