package clause

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

// Condition 是表达式的规范化结构表示，便于统一遍历与改写。
type Condition struct {
	Column string
	Op     Operator
	Values []any // 单值比较长度 1；IN 为多个；LIKE 为单个字符串
}

// AsCondition 将 Expression 转换成 Condition。逻辑组合等不支持的类型 ok=false。
func AsCondition(expr Expression) (*Condition, bool) {
	switch e := expr.(type) {
	case Eq:
		return &Condition{e.Column, OpEQ, []any{e.Value}}, true
	case Neq:
		return &Condition{e.Column, OpNEQ, []any{e.Value}}, true
	case Gt:
		return &Condition{e.Column, OpGT, []any{e.Value}}, true
	case Gte:
		return &Condition{e.Column, OpGTE, []any{e.Value}}, true
	case Lt:
		return &Condition{e.Column, OpLT, []any{e.Value}}, true
	case Lte:
		return &Condition{e.Column, OpLTE, []any{e.Value}}, true
	case Like:
		return &Condition{e.Column, OpLIKE, []any{e.Value}}, true
	case IN:
		return &Condition{e.Column, OpIN, e.Values}, true
	}
	return nil, false
}

func (c *Condition) ToExpr() Expression {

	if c == nil {
		return nil
	}

	if c.Column == "" {
		return nil
	}

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

// Map 遍历表达式列表，并生成新的表达式列表
//
// mapper 为表达式遍历函数，返回 nil 表示移除该表达式
//   - e Expression 原表达式
//   - c *Condition 解析出的 Condition 指针，当无法解析为 Condition 时，c 为 nil，如 And 等逻辑组合表达式
//
// 返回新的 Where 对象
func (w Where) Map(mapper func(e Expression, c *Condition) Expression) Where {
	exprs := mapExprs(w.Exprs, mapper)
	return Where{Exprs: exprs}
}

func mapExprs(exprs []Expression, mapper func(e Expression, c *Condition) Expression) []Expression {
	result := make([]Expression, 0, len(exprs))
	for _, expr := range exprs {
		c, ok := AsCondition(expr)
		if !ok {

			newExp := mapper(expr, nil)
			if newExp == nil {
				continue
			}

			switch e := newExp.(type) {
			case AndExpr:
				exprs := mapExprs(e.Exprs, mapper)
				if expr := And(exprs...); expr != nil {
					result = append(result, expr)
				}
			case OrExpr:
				exprs := mapExprs(e.Exprs, mapper)
				if expr := Or(exprs...); expr != nil {
					result = append(result, expr)
				}
			case NotExpr:
				exprs := mapExprs(e.Exprs, mapper)
				if expr := Not(exprs...); expr != nil {
					result = append(result, expr)
				}
			}
			continue
		}
		if e := mapper(expr, c); e != nil {
			result = append(result, e)
		}
	}
	return result
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
