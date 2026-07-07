package clause

import "strings"

// 连接字符串常量，用于 Build 时连接多个表达式。
const (
	AndWithSpace = " AND "
	OrWithSpace  = " OR "
)

// Operator 表示 SQL 比较操作符。
// 在 MapColumn 等方法中用于标识表达式类型。
type Operator string

// 支持的 SQL 比较操作符。
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
			if value, ok := e.Value.(string); ok {
				prefix, suffix := "", ""
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
			}
			return e

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

// Condition 是表达式的规范化结构表示，便于统一遍历与改写。
type Condition struct {
	Column string
	Op     Operator
	Values []any // 单值比较长度 1；IN 为多个；LIKE 为单个字符串
}

// toCondition 将 Expression 转换成 Condition。逻辑组合等不支持的类型 ok=false。
func toCondition(expr Expression) (Condition, bool) {
	switch e := expr.(type) {
	case Eq:
		return Condition{e.Column, OpEQ, []any{e.Value}}, true
	case Neq:
		return Condition{e.Column, OpNEQ, []any{e.Value}}, true
	case Gt:
		return Condition{e.Column, OpGT, []any{e.Value}}, true
	case Gte:
		return Condition{e.Column, OpGTE, []any{e.Value}}, true
	case Lt:
		return Condition{e.Column, OpLT, []any{e.Value}}, true
	case Lte:
		return Condition{e.Column, OpLTE, []any{e.Value}}, true
	case Like:
		return Condition{e.Column, OpLIKE, []any{e.Value}}, true
	case IN:
		return Condition{e.Column, OpIN, e.Values}, true
	}
	return Condition{}, false
}

// toExpression 由 Condition 重建 Expression。
func toExpression(c Condition) Expression {
	switch c.Op {
	case OpEQ:
		return Eq{Column: c.Column, Value: c.Values[0]}
	case OpNEQ:
		return Neq{Column: c.Column, Value: c.Values[0]}
	case OpGT:
		return Gt{Column: c.Column, Value: c.Values[0]}
	case OpGTE:
		return Gte{Column: c.Column, Value: c.Values[0]}
	case OpLT:
		return Lt{Column: c.Column, Value: c.Values[0]}
	case OpLTE:
		return Lte{Column: c.Column, Value: c.Values[0]}
	case OpLIKE:
		return Like{Column: c.Column, Value: c.Values[0]}
	case OpIN:
		return IN{Column: c.Column, Values: c.Values}
	}
	return nil
}

func (w Where) MapCondition(mapper func(Condition) Condition) Where {
	return w.Map(func(expr Expression) Expression {
		c, ok := toCondition(expr)
		if !ok {
			return expr // 逻辑组合等不支持的，原样保留
		}
		c = mapper(c)
		if c.Column == "" {
			return nil // 移除
		}
		return toExpression(c)
	})
}

// And 将多个表达式用 AND 逻辑组合。
// 如果只有一个非 OrExpr 表达式，则直接返回该表达式（去掉多余的 AND 包裹）。
// 如果只有一个 OrExpr 表达式，则仍用 AND 包裹以保持逻辑清晰。
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

// AndExpr 表示 AND 逻辑组合表达式。
// 多个子表达式之间用 AND 连接。
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

// Or 将多个表达式用 OR 逻辑组合。
func Or(exprs ...Expression) Expression {
	if len(exprs) == 0 {
		return nil
	}
	return OrExpr{Exprs: exprs}
}

// OrExpr 表示 OR 逻辑组合表达式。
// 多个子表达式之间用 OR 连接。
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

// Not 对表达式取反（NOT）。
// 如果传入单个 AndExpr，会解开其内部的子表达式后取反。
// 对于实现了 NegationExpressionBuilder 接口的表达式，
// 在构建时会调用 NegationBuild 方法生成更自然的否定形式。
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

// NotExpr 表示 NOT 取反表达式。
// 构建时会优先使用 NegationExpressionBuilder 接口生成自然否定形式。
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
