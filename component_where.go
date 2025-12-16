package query

import "github.com/epkgs/query/clause"

type Wherer interface {
	WhereExpr() clause.Where
	Where(field any, args ...any) Wherer
	OrWhere(field any, args ...any) Wherer
	Not(field any, args ...any) Wherer
}

type where[P errorRecorder] struct {
	Parent P
	Value  clause.Where
}

func (p *where[P]) WhereExpr() clause.Where {
	return p.Value
}

// Where 添加WHERE条件到当前查询
// 参数:
//   - field: 字段名、表达式或表达式数组
//   - args: 条件参数，格式根据field类型而定
//
// 返回值:
//   - 当前实例，支持链式调用
//
// 示例:
//   - q.Where("name", "John")             // name = John
//   - q.Where("age", ">", 18)              // age > 18
//   - q.Where(clause.Eq{Column: "name", Value: "John"})  // 使用clause.Expression
//   - q.Where(func(w Wherer) Wherer {  w.Where("name", "John"); return w }) // Closure
func (p *where[P]) Where(field any, args ...any) P {
	expressions, err := buildCondition(field, args...)
	if err != nil {
		p.Parent.setError(err)
		return p.Parent
	}
	p.Value.Merge(clause.Where{Exprs: expressions})
	return p.Parent
}

// OrWhere 添加OR WHERE条件到当前查询
// 参数:
//   - field: 字段名、表达式或表达式数组
//   - args: 条件参数，格式根据field类型而定
//
// 返回值:
//   - 当前实例，支持链式调用
//
// 示例:
//   - q.OrWhere("name", "John")             // OR name = John
//   - q.OrWhere("age", ">", 18)              // OR age > 18
//   - q.OrWhere(clause.Eq{Column: "name", Value: "John"})  // 使用clause.Expression
//   - q.OrWhere(func(w Wherer) Wherer {  w.Where("name", "John"); return w }) // Closure
func (p *where[P]) OrWhere(field any, args ...any) P {
	expressions, err := buildCondition(field, args...)
	if err != nil {
		p.Parent.setError(err)
		return p.Parent
	}
	if len(expressions) > 0 {
		p.Value.Merge(clause.Where{Exprs: []clause.Expression{clause.Or(expressions...)}})
	}
	return p.Parent
}

// Not 添加NOT条件到当前查询
// 参数:
//   - field: 字段名、表达式或表达式数组
//   - args: 条件参数，格式根据field类型而定
//
// 返回值:
//   - 当前实例，支持链式调用
//
// 示例:
//   - q.Not("name", "John")             // NOT name = John
//   - q.Not("age", ">", 18)              // NOT age > 18
//   - q.Not(clause.Eq{Column: "name", Value: "John"})  // 使用clause.Expression
//   - q.Not(func(w Wherer) Wherer {  w.Where("name", "John"); return w }) // Closure
func (p *where[P]) Not(field any, args ...any) P {
	expressions, err := buildCondition(field, args...)
	if err != nil {
		p.Parent.setError(err)
		return p.Parent
	}
	p.Value.Merge(clause.Where{Exprs: []clause.Expression{clause.Not(expressions...)}})
	return p.Parent
}

// buildWhereClause 构建WHERE子句
func (w *where[P]) Build(builder clause.Builder) {
	w.Value.Build(builder)
}

func NewWhereBuilder() *WhereBuilder {
	return &WhereBuilder{
		where: clause.Where{},
	}
}

var _ Wherer = (*WhereBuilder)(nil)

// WhereBuilder 实现 Wherer 接口
type WhereBuilder struct {
	where clause.Where
	Error error
}

func (w *WhereBuilder) WhereExpr() clause.Where {
	return w.where
}

func (w *WhereBuilder) Where(field any, args ...any) Wherer {
	expressions, err := buildCondition(field, args...)
	if err != nil {
		w.Error = err
		return w
	}
	w.where.Merge(clause.Where{Exprs: expressions})
	return w
}

func (w *WhereBuilder) OrWhere(field any, args ...any) Wherer {
	expressions, err := buildCondition(field, args...)
	if err != nil {
		w.Error = err
		return w
	}
	if len(expressions) > 0 {
		w.where.Merge(clause.Where{Exprs: []clause.Expression{clause.Or(expressions...)}})
	}
	return w
}

func (w *WhereBuilder) Not(field any, args ...any) Wherer {
	expressions, err := buildCondition(field, args...)
	if err != nil {
		w.Error = err
		return w
	}
	w.where.Merge(clause.Where{Exprs: []clause.Expression{clause.Not(expressions...)}})
	return w
}

// buildCondition 构建条件表达式
// 参数:
//   - column: 字段名、表达式或表达式数组
//   - args: 条件参数，格式根据column类型而定
//
// 返回值:
//   - []clause.Expression: 构建的条件表达式数组
//   - error: 构建过程中遇到的错误（如果有）
//
// 功能说明:
//   - 根据输入类型构建不同的条件表达式
//   - 支持字符串字段名 + 操作符 + 值的格式
//   - 支持直接传入clause.Expression接口实现
//   - 支持传入[]clause.Expression数组
//   - 支持IN条件（当第二个参数是数组时）
func buildCondition(column any, args ...any) ([]clause.Expression, error) {

	switch c := column.(type) {
	case func(Wherer) Wherer:
		// 创建whereBuilder实例
		builder := NewWhereBuilder()
		// 调用闭包函数
		result := c(builder).WhereExpr().Exprs
		// 用And包裹以添加括号
		return []clause.Expression{clause.And(result...)}, nil
	case clause.Where:
		return c.Exprs, nil
	case clause.AndExpr:
		// 如果是AndExpr类型，将其作为单个表达式返回
		return []clause.Expression{c}, nil
	case clause.OrExpr:
		// 如果是OrExpr类型，将其作为单个表达式返回
		return []clause.Expression{c}, nil
	case []clause.Expression:
		// 如果是Expression数组，直接返回
		return c, nil
	case clause.Expression:
		// 如果是单个Expression接口实现，将其作为数组返回
		return []clause.Expression{c}, nil
	case string:
		// 如果是字符串字段名，根据参数数量构建不同条件
		if len(args) == 0 {
			// 只有一个参数，构建 column = value 条件（column为空）
			return []clause.Expression{clause.Eq{Column: "", Value: c}}, nil
		}
		if len(args) == 1 {
			// 两个参数，判断第二个参数是否为数组
			if arr, ok := args[0].([]interface{}); ok {
				// 如果是数组，构建IN条件
				return []clause.Expression{clause.IN{Column: c, Values: arr}}, nil
			} else {
				// 如果不是数组，构建=条件
				return []clause.Expression{clause.Eq{Column: c, Value: args[0]}}, nil
			}
		}

		// 三个或更多参数，解析为 column operator value 格式
		op, ok := args[0].(string)
		if !ok {
			// 操作符必须是字符串类型
			return nil, ErrInvalidOperator
		}

		// 根据操作符构建不同的比较条件
		switch op {
		case "=":
			return []clause.Expression{clause.Eq{Column: c, Value: args[1]}}, nil
		case "!=":
			return []clause.Expression{clause.Neq{Column: c, Value: args[1]}}, nil
		case ">":
			return []clause.Expression{clause.Gt{Column: c, Value: args[1]}}, nil
		case ">=":
			return []clause.Expression{clause.Gte{Column: c, Value: args[1]}}, nil
		case "<":
			return []clause.Expression{clause.Lt{Column: c, Value: args[1]}}, nil
		case "<=":
			return []clause.Expression{clause.Lte{Column: c, Value: args[1]}}, nil
		case "LIKE":
			return []clause.Expression{clause.Like{Column: c, Value: args[1]}}, nil
		case "IN":
			return []clause.Expression{clause.IN{Column: c, Values: args[1].([]interface{})}}, nil
		default:
			// 不支持的操作符
			return nil, ErrInvalidOperator
		}

	}

	// 无法识别的column类型
	return nil, ErrInvalidCondition
}
