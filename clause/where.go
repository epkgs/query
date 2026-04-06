package clause

import "strings"

const (
	AndWithSpace = " AND "
	OrWithSpace  = " OR "
)

type Operator string

const (
	OpEQ   Operator = "="
	OpNEQ  Operator = "!="
	OpGT   Operator = ">"
	OpGTE  Operator = ">="
	OpLT   Operator = "<"
	OpLTE  Operator = "<="
	OpLIKE Operator = "LIKE"
	OpIN   Operator = "IN"
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

// Map 遍历表达式列表，并生成新的表达式列表
//
// mapper 为表达式遍历函数，返回 nil 表示移除该表达式
func (w Where) Map(mapper func(expr Expression) Expression) Where {
	exprs := mapExpressions(w.Exprs, mapper)
	return Where{Exprs: exprs}
}

func mapExpressions(exprs []Expression, walkers ...func(Expression) Expression) []Expression {

	copied := make([]Expression, len(exprs))
	copy(copied, exprs)

	if len(walkers) == 0 {
		return copied
	}

	result := []Expression{}
	for _, exp := range copied {
		switch e := exp.(type) {
		case AndExpr:
			e.Exprs = mapExpressions(e.Exprs, walkers...)
			result = append(result, e)

		case OrExpr:
			e.Exprs = mapExpressions(e.Exprs, walkers...)
			result = append(result, e)

		case NotExpr:
			e.Exprs = mapExpressions(e.Exprs, walkers...)
			result = append(result, e)

		default:

			for _, walk := range walkers {
				if walk == nil {
					continue
				}
				if e == nil {
					break
				}
				e = walk(e)
			}

			if e != nil {
				result = append(result, e)
			}
		}

	}
	return result
}

// MapColumn 遍历表达式列表，并生成新的表达式列表
//
// mapper 为表达式遍历函数，返回空字符串表示移除该表达式
func (w Where) MapColumn(mapper func(column string, op Operator, value any) (string, any)) Where {
	return w.Map(func(expr Expression) Expression {

		if expr == nil {
			return nil
		}

		switch e := expr.(type) {
		case Eq:
			if e.Column, e.Value = mapper(e.Column, OpEQ, e.Value); e.Column != "" {
				return e
			}
			return nil
		case Neq:
			if e.Column, e.Value = mapper(e.Column, OpNEQ, e.Value); e.Column != "" {
				return e
			}
			return nil
		case Gt:
			if e.Column, e.Value = mapper(e.Column, OpGT, e.Value); e.Column != "" {
				return e
			}
			return nil
		case Gte:
			if e.Column, e.Value = mapper(e.Column, OpGTE, e.Value); e.Column != "" {
				return e
			}
			return nil
		case Lt:
			if e.Column, e.Value = mapper(e.Column, OpLT, e.Value); e.Column != "" {
				return e
			}
			return nil
		case Lte:
			if e.Column, e.Value = mapper(e.Column, OpLTE, e.Value); e.Column != "" {
				return e
			}
			return nil
		case Like:
			prefix, suffix := "", ""
			value := e.Value.(string)
			if strings.HasPrefix(value, "%") {
				prefix = "%"
				value = strings.TrimPrefix(value, "%")
			}
			if strings.HasSuffix(value, "%") {
				suffix = "%"
				value = strings.TrimSuffix(value, "%")
			}
			if col, val := mapper(e.Column, OpLIKE, value); col != "" {
				e.Column = col
				e.Value = prefix + val.(string) + suffix
				return e
			}
			return nil
		case IN:
			values := e.Values
			column := e.Column
			for i, v := range values {
				column, values[i] = mapper(e.Column, OpIN, v)
				if column == "" {
					return nil
				}
			}
			e.Column = column
			e.Values = values
			return e
		}
		return expr
	})
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
