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

// LogicalOperator 表示逻辑运算符。
type LogicalOperator int

// 支持的逻辑运算符。
const (
	LogicAnd LogicalOperator = iota
	LogicOr
	LogicNot
)

// LogicalExpression 表示逻辑组合表达式。
// AndExpr、OrExpr、NotExpr 等类型实现此接口。
type LogicalExpression interface {
	Expression
	Operator() LogicalOperator
	SubExprs() []Expression
	logicalExpr() // 密封方法，仅 clause 包内实现
}

// Where where clause
type Where struct {
	Exprs []Expression
}

// Build build where clause
func (w Where) Build(builder Builder) {
	exprs := w.Exprs
	if len(exprs) == 1 {
		if logical, ok := exprs[0].(LogicalExpression); ok && logical.Operator() == LogicAnd {
			exprs = logical.SubExprs()
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
//   - e Expression 原表达式
//
// 返回新的 Where 对象
func (w Where) Map(mapper func(e Expression) Expression) Where {
	exprs := mapExprs(w.Exprs, mapper)
	return Where{Exprs: exprs}
}

func mapExprs(exprs []Expression, mapper func(e Expression) Expression) []Expression {
	result := make([]Expression, 0, len(exprs))
	for _, expr := range exprs {
		newExp := mapper(expr)
		if newExp == nil {
			continue
		}

		if logical, ok := newExp.(LogicalExpression); ok {
			subExprs := mapExprs(logical.SubExprs(), mapper)
			switch logical.Operator() {
			case LogicAnd:
				if expr := And(subExprs...); expr != nil {
					result = append(result, expr)
				}
			case LogicOr:
				if expr := Or(subExprs...); expr != nil {
					result = append(result, expr)
				}
			case LogicNot:
				if expr := Not(subExprs...); expr != nil {
					result = append(result, expr)
				}
			}
			continue
		}

		result = append(result, newExp)
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
		if logical, ok := exprs[0].(LogicalExpression); !ok || logical.Operator() != LogicOr {
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

func (and AndExpr) Operator() LogicalOperator { return LogicAnd }
func (and AndExpr) SubExprs() []Expression    { return and.Exprs }
func (and AndExpr) logicalExpr()              {}

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

func (or OrExpr) Operator() LogicalOperator { return LogicOr }
func (or OrExpr) SubExprs() []Expression    { return or.Exprs }
func (or OrExpr) logicalExpr()              {}

// Not 对表达式取反（NOT）。
// 如果传入单个 AndExpr，会解开其内部的子表达式后取反。
// 对于实现了 NegationExpressionBuilder 接口的表达式，
// 在构建时会调用 NegationBuild 方法生成更自然的否定形式。
func Not(exprs ...Expression) Expression {
	if len(exprs) == 0 {
		return nil
	}
	if len(exprs) == 1 {
		if logical, ok := exprs[0].(LogicalExpression); ok && logical.Operator() == LogicAnd {
			exprs = logical.SubExprs()
		}
	}
	return NotExpr{Exprs: exprs}
}

// NotExpr 表示 NOT 取反表达式。
// 构建时会优先使用 NegationExpressionBuilder 接口生成自然否定形式。
type NotExpr struct {
	Exprs []Expression
}

func (not NotExpr) Operator() LogicalOperator { return LogicNot }
func (not NotExpr) SubExprs() []Expression    { return not.Exprs }
func (not NotExpr) logicalExpr()              {}

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
				if logical, ok := c.(LogicalExpression); ok && logical.Operator() == LogicOr {
					builder.WriteString(" OR ")
				} else {
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
			if logical, ok := expr.(LogicalExpression); ok && logical.Operator() == LogicOr && len(logical.SubExprs()) == 1 {
				builder.WriteString(OrWithSpace)
			} else {
				builder.WriteString(joinCond)
			}
		}

		expr.Build(builder)
	}
}
