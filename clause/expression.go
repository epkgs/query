package clause

import "reflect"

// Expression 是查询表达式的抽象接口。
// 所有查询组件（比较条件、逻辑组合、排序、分页等）都实现此接口，
// 通过 Build 方法将自身写入 Builder。
type Expression interface {
	Build(builder Builder)
}

// ComparisonExpression 表示比较表达式（叶子谓词）。
// Eq、Neq、Gt、Gte、Lt、Lte、Like、IN 等类型实现此接口。
type ComparisonExpression interface {
	Expression
	Operator() Operator  // 返回比较操作符
	Column() string      // 返回比较字段
	Value() any          // 返回比较值；IN 返回 Vals 切片
	comparisonExpr()     // 密封方法，仅 clause 包内实现
}

// NegationExpressionBuilder 表示支持否定构建的表达式。
// 当表达式用于 NOT 上下文中时（如 NotExpr），
// 如果表达式实现了此接口，则调用 NegationBuild 而非 Build，
// 从而生成更自然的否定 SQL（例如将 = 转换为 <>）。
type NegationExpressionBuilder interface {
	NegationBuild(builder Builder)
}

// IN Whether a value is within a set of values
type IN struct {
	Col  string
	Vals []interface{}
}

func (in IN) Build(builder Builder) {
	builder.WriteQuoted(in.Col)

	switch len(in.Vals) {
	case 0:
		builder.WriteString(" IN (NULL)")
	case 1:
		if _, ok := in.Vals[0].([]interface{}); !ok {
			builder.WriteString(" = ")
			builder.AddVar(builder, in.Vals[0])
			break
		}

		fallthrough
	default:
		builder.WriteString(" IN (")
		builder.AddVar(builder, in.Vals...)
		builder.WriteByte(')')
	}
}

func (in IN) NegationBuild(builder Builder) {
	builder.WriteQuoted(in.Col)
	switch len(in.Vals) {
	case 0:
		builder.WriteString(" IS NOT NULL")
	case 1:
		if _, ok := in.Vals[0].([]interface{}); !ok {
			builder.WriteString(" <> ")
			builder.AddVar(builder, in.Vals[0])
			break
		}

		fallthrough
	default:
		builder.WriteString(" NOT IN (")
		builder.AddVar(builder, in.Vals...)
		builder.WriteByte(')')
	}
}

func (in IN) comparisonExpr() {}
func (in IN) Operator() Operator { return OpIN }
func (in IN) Column() string     { return in.Col }
func (in IN) Value() any         { return in.Vals }

// Eq equal to for where
type Eq struct {
	Col string
	Val interface{}
}

func buildEqArray[T any](builder Builder, values []T) {
	if len(values) == 0 {
		builder.WriteString(" IN (NULL)")
	} else {
		builder.WriteString(" IN (")
		for i := 0; i < len(values); i++ {
			if i > 0 {
				builder.WriteByte(',')
			}
			builder.AddVar(builder, values[i])
		}
		builder.WriteByte(')')
	}
}

func (eq Eq) Build(builder Builder) {
	builder.WriteQuoted(eq.Col)

	switch val := eq.Val.(type) {
	case []string:
		buildEqArray(builder, val)
	case []int:
		buildEqArray(builder, val)
	case []int32:
		buildEqArray(builder, val)
	case []int64:
		buildEqArray(builder, val)
	case []uint:
		buildEqArray(builder, val)
	case []uint32:
		buildEqArray(builder, val)
	case []uint64:
		buildEqArray(builder, val)
	case []interface{}:
		buildEqArray(builder, val)
	default:
		if eqNil(eq.Val) {
			builder.WriteString(" IS NULL")
		} else {
			builder.WriteString(" = ")
			builder.AddVar(builder, eq.Val)
		}
	}
}

func (eq Eq) NegationBuild(builder Builder) {
	Neq(eq).Build(builder)
}

func (eq Eq) comparisonExpr() {}
func (eq Eq) Operator() Operator { return OpEQ }
func (eq Eq) Column() string     { return eq.Col }
func (eq Eq) Value() any         { return eq.Val }

// Neq not equal to for where
type Neq Eq

func buildNeqArray[T any](builder Builder, values []T) {
	if len(values) == 0 {
		builder.WriteString(" NOT IN (NULL)")
	} else {
		builder.WriteString(" NOT IN (")
		for i := 0; i < len(values); i++ {
			if i > 0 {
				builder.WriteByte(',')
			}
			builder.AddVar(builder, values[i])
		}
		builder.WriteByte(')')
	}
}

func (neq Neq) Build(builder Builder) {
	builder.WriteQuoted(neq.Col)

	switch val := neq.Val.(type) {
	case []string:
		buildNeqArray(builder, val)
	case []int:
		buildNeqArray(builder, val)
	case []int32:
		buildNeqArray(builder, val)
	case []int64:
		buildNeqArray(builder, val)
	case []uint:
		buildNeqArray(builder, val)
	case []uint32:
		buildNeqArray(builder, val)
	case []uint64:
		buildNeqArray(builder, val)
	case []interface{}:
		buildNeqArray(builder, val)
	default:
		if eqNil(neq.Val) {
			builder.WriteString(" IS NOT NULL")
		} else {
			builder.WriteString(" <> ")
			builder.AddVar(builder, neq.Val)
		}
	}
}

func (neq Neq) NegationBuild(builder Builder) {
	Eq(neq).Build(builder)
}
func (neq Neq) comparisonExpr()  {}
func (neq Neq) Operator() Operator { return OpNEQ }
func (neq Neq) Column() string     { return neq.Col }
func (neq Neq) Value() any         { return neq.Val }

// Gt greater than for where
type Gt Eq

func (gt Gt) Build(builder Builder) {
	builder.WriteQuoted(gt.Col)
	builder.WriteString(" > ")
	builder.AddVar(builder, gt.Val)
}

func (gt Gt) NegationBuild(builder Builder) {
	Lte(gt).Build(builder)
}
func (gt Gt) comparisonExpr()  {}
func (gt Gt) Operator() Operator { return OpGT }
func (gt Gt) Column() string     { return gt.Col }
func (gt Gt) Value() any         { return gt.Val }

// Gte greater than or equal to for where
type Gte Eq

func (gte Gte) Build(builder Builder) {
	builder.WriteQuoted(gte.Col)
	builder.WriteString(" >= ")
	builder.AddVar(builder, gte.Val)
}

func (gte Gte) NegationBuild(builder Builder) {
	Lt(gte).Build(builder)
}
func (gte Gte) comparisonExpr()  {}
func (gte Gte) Operator() Operator { return OpGTE }
func (gte Gte) Column() string     { return gte.Col }
func (gte Gte) Value() any         { return gte.Val }

// Lt less than for where
type Lt Eq

func (lt Lt) Build(builder Builder) {
	builder.WriteQuoted(lt.Col)
	builder.WriteString(" < ")
	builder.AddVar(builder, lt.Val)
}

func (lt Lt) NegationBuild(builder Builder) {
	Gte(lt).Build(builder)
}
func (lt Lt) comparisonExpr()  {}
func (lt Lt) Operator() Operator { return OpLT }
func (lt Lt) Column() string     { return lt.Col }
func (lt Lt) Value() any         { return lt.Val }

// Lte less than or equal to for where
type Lte Eq

func (lte Lte) Build(builder Builder) {
	builder.WriteQuoted(lte.Col)
	builder.WriteString(" <= ")
	builder.AddVar(builder, lte.Val)
}

func (lte Lte) NegationBuild(builder Builder) {
	Gt(lte).Build(builder)
}
func (lte Lte) comparisonExpr()  {}
func (lte Lte) Operator() Operator { return OpLTE }
func (lte Lte) Column() string     { return lte.Col }
func (lte Lte) Value() any         { return lte.Val }

// Like whether string matches regular expression
type Like Eq

func (like Like) Build(builder Builder) {
	builder.WriteQuoted(like.Col)
	builder.WriteString(" LIKE ")
	builder.AddVar(builder, like.Val)
}

func (like Like) NegationBuild(builder Builder) {
	builder.WriteQuoted(like.Col)
	builder.WriteString(" NOT LIKE ")
	builder.AddVar(builder, like.Val)
}
func (like Like) comparisonExpr()  {}
func (like Like) Operator() Operator { return OpLIKE }
func (like Like) Column() string     { return like.Col }
func (like Like) Value() any         { return like.Val }

func eqNil(value interface{}) bool {
	if valuer, ok := value.(Valuer); ok && !eqNilReflect(valuer) {
		value, _ = valuer.Value()
	}

	return value == nil || eqNilReflect(value)
}

func eqNilReflect(value interface{}) bool {
	reflectValue := reflect.ValueOf(value)
	return reflectValue.Kind() == reflect.Ptr && reflectValue.IsNil()
}
