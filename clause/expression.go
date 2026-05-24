package clause

import "reflect"

// Expression 是查询表达式的抽象接口。
// 所有查询组件（比较条件、逻辑组合、排序、分页等）都实现此接口，
// 通过 Build 方法将自身写入 Builder。
type Expression interface {
	Build(builder Builder)
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
	Column string
	Values []interface{}
}

func (in IN) Build(builder Builder) {
	builder.WriteQuoted(in.Column)

	switch len(in.Values) {
	case 0:
		builder.WriteString(" IN (NULL)")
	case 1:
		if _, ok := in.Values[0].([]interface{}); !ok {
			builder.WriteString(" = ")
			builder.AddVar(builder, in.Values[0])
			break
		}

		fallthrough
	default:
		builder.WriteString(" IN (")
		builder.AddVar(builder, in.Values...)
		builder.WriteByte(')')
	}
}

func (in IN) NegationBuild(builder Builder) {
	builder.WriteQuoted(in.Column)
	switch len(in.Values) {
	case 0:
		builder.WriteString(" IS NOT NULL")
	case 1:
		if _, ok := in.Values[0].([]interface{}); !ok {
			builder.WriteString(" <> ")
			builder.AddVar(builder, in.Values[0])
			break
		}

		fallthrough
	default:
		builder.WriteString(" NOT IN (")
		builder.AddVar(builder, in.Values...)
		builder.WriteByte(')')
	}
}

// Eq equal to for where
type Eq struct {
	Column string
	Value  interface{}
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
	builder.WriteQuoted(eq.Column)

	switch val := eq.Value.(type) {
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
		if eqNil(eq.Value) {
			builder.WriteString(" IS NULL")
		} else {
			builder.WriteString(" = ")
			builder.AddVar(builder, eq.Value)
		}
	}
}

func (eq Eq) NegationBuild(builder Builder) {
	Neq(eq).Build(builder)
}

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
	builder.WriteQuoted(neq.Column)

	switch val := neq.Value.(type) {
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
		if eqNil(neq.Value) {
			builder.WriteString(" IS NOT NULL")
		} else {
			builder.WriteString(" <> ")
			builder.AddVar(builder, neq.Value)
		}
	}
}

func (neq Neq) NegationBuild(builder Builder) {
	Eq(neq).Build(builder)
}

// Gt greater than for where
type Gt Eq

func (gt Gt) Build(builder Builder) {
	builder.WriteQuoted(gt.Column)
	builder.WriteString(" > ")
	builder.AddVar(builder, gt.Value)
}

func (gt Gt) NegationBuild(builder Builder) {
	Lte(gt).Build(builder)
}

// Gte greater than or equal to for where
type Gte Eq

func (gte Gte) Build(builder Builder) {
	builder.WriteQuoted(gte.Column)
	builder.WriteString(" >= ")
	builder.AddVar(builder, gte.Value)
}

func (gte Gte) NegationBuild(builder Builder) {
	Lt(gte).Build(builder)
}

// Lt less than for where
type Lt Eq

func (lt Lt) Build(builder Builder) {
	builder.WriteQuoted(lt.Column)
	builder.WriteString(" < ")
	builder.AddVar(builder, lt.Value)
}

func (lt Lt) NegationBuild(builder Builder) {
	Gte(lt).Build(builder)
}

// Lte less than or equal to for where
type Lte Eq

func (lte Lte) Build(builder Builder) {
	builder.WriteQuoted(lte.Column)
	builder.WriteString(" <= ")
	builder.AddVar(builder, lte.Value)
}

func (lte Lte) NegationBuild(builder Builder) {
	Gt(lte).Build(builder)
}

// Like whether string matches regular expression
type Like Eq

func (like Like) Build(builder Builder) {
	builder.WriteQuoted(like.Column)
	builder.WriteString(" LIKE ")
	builder.AddVar(builder, like.Value)
}

func (like Like) NegationBuild(builder Builder) {
	builder.WriteQuoted(like.Column)
	builder.WriteString(" NOT LIKE ")
	builder.AddVar(builder, like.Value)
}

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
