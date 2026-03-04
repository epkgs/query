package query

import (
	"reflect"

	"github.com/epkgs/query/clause"
)

type genericWherer[Q any] interface {
	WhereExpr() clause.Where
	Where(field any, args ...any) Q
	OrWhere(field any, args ...any) Q
	NotWhere(field any, args ...any) Q
	And(querys ...Q) Q
	Eq(column string, value any) Q
	Gt(column string, value any) Q
	Gte(column string, value any) Q
	In(column string, values ...any) Q
	Like(column string, value any) Q
	Lt(column string, value any) Q
	Lte(column string, value any) Q
	Neq(column string, value any) Q
	Not(query Q) Q
	Or(querys ...Q) Q
}

var _ genericWherer[*Query] = (*where[*Query])(nil)
var _ clause.Expression = (*where[*Query])(nil)

type whereExprExporter interface {
	WhereExpr() clause.Where
}

type whereQuerier interface {
	errorRecorder
	whereExprExporter
}

type where[Q whereQuerier] struct {
	Parent Q
	Value  clause.Where
}

func (w *where[Q]) WhereExpr() clause.Where {
	return w.Value
}

// Where 添加WHERE条件到当前查询
//
// Deprecated: 使用 Eq, Neq, Gt 等方法替代，例如：q.Eq("name", "John")
//
// 参数:
//   - field: 字段名、表达式或表达式数组
//   - args: 条件参数，格式根据field类型而定
//
// 返回值:
//   - 当前实例，支持链式调用
//
// 示例:
//   - q.Where("name", "John")
//   - q.Where("age", ">", 18)
//   - q.Where(clause.Eq{Column: "name", Value: "John"})
//   - q.Where(func(w Wherer) Wherer {  w.Where("name", "John"); return w.Parent })
func (w *where[Q]) Where(field any, args ...any) Q {
	expressions, err := buildCondition(field, args...)
	if err != nil {
		w.Parent.setError(err)
		return w.Parent
	}
	w.Value.Merge(clause.Where{Exprs: expressions})
	return w.Parent
}

// OrWhere 添加OR WHERE条件到当前查询
//
// Deprecated: 使用 Or() 函数替代，例如：Or(q.Eq("name", "John"), q.Eq("name", "Jane"))
//
// 参数:
//   - field: 字段名、表达式或表达式数组
//   - args: 条件参数，格式根据field类型而定
//
// 返回值:
//   - 当前实例，支持链式调用
//
// 示例:
//   - q.OrWhere("name", "John")
//   - q.OrWhere("age", ">", 18)
//   - q.OrWhere(clause.Eq{Column: "name", Value: "John"})
//   - q.OrWhere(func(w Wherer) Wherer {  w.Where("name", "John"); return w.Parent })
func (w *where[Q]) OrWhere(field any, args ...any) Q {
	expressions, err := buildCondition(field, args...)
	if err != nil {
		w.Parent.setError(err)
		return w.Parent
	}
	if len(expressions) > 0 {
		w.Value.Merge(clause.Where{Exprs: []clause.Expression{clause.Or(expressions...)}})
	}
	return w.Parent
}

// Not 添加NOT条件到当前查询
//
// Deprecated: 使用 Not() 函数替代，例如：Not(q.Eq("name", "John"))
//
// 参数:
//   - field: 字段名、表达式或表达式数组
//   - args: 条件参数，格式根据field类型而定
//
// 返回值:
//   - 当前实例，支持链式调用
//
// 示例:
//   - q.Not("name", "John")
//   - q.Not("age", ">", 18)
//   - q.Not(clause.Eq{Column: "name", Value: "John"})
//   - q.Not(func(w Wherer) Wherer {  w.Where("name", "John"); return w.Parent })
func (w *where[Q]) NotWhere(field any, args ...any) Q {
	expressions, err := buildCondition(field, args...)
	if err != nil {
		w.Parent.setError(err)
		return w.Parent
	}
	w.Value.Merge(clause.Where{Exprs: []clause.Expression{clause.Not(expressions...)}})
	return w.Parent
}

// buildWhereClause 构建WHERE子句
func (w *where[Q]) Build(builder clause.Builder) {
	w.Value.Build(builder)
}

type Wherer interface {
	WhereExpr() clause.Where
	Where(field any, args ...any) Wherer
	OrWhere(field any, args ...any) Wherer
	NotWhere(field any, args ...any) Wherer
}

var _ Wherer = (*whereBuilder)(nil)

// whereBuilder 实现 Wherer 接口
type whereBuilder struct {
	where clause.Where
	Error error
}

func (w *whereBuilder) WhereExpr() clause.Where {
	return w.where
}

func (w *whereBuilder) Where(field any, args ...any) Wherer {
	expressions, err := buildCondition(field, args...)
	if err != nil {
		w.Error = err
		return w
	}
	w.where.Merge(clause.Where{Exprs: expressions})
	return w
}

func (w *whereBuilder) OrWhere(field any, args ...any) Wherer {
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

func (w *whereBuilder) NotWhere(field any, args ...any) Wherer {
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
		builder := &whereBuilder{where: clause.Where{}}
		// 调用闭包函数
		exprs := c(builder).WhereExpr().Exprs
		if len(exprs) == 0 {
			return exprs, nil
		}
		// 用And包裹以添加括号
		return []clause.Expression{clause.And(exprs...)}, nil
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
		case "LIKE", "like":
			return []clause.Expression{clause.Like{Column: c, Value: args[1]}}, nil
		case "IN", "in":
			return []clause.Expression{clause.IN{Column: c, Values: toAnySlice(args[1])}}, nil
		default:
			// 不支持的操作符
			return nil, ErrInvalidOperator
		}

	}

	// 无法识别的column类型
	return nil, ErrInvalidCondition
}

func toAnySlice(value any) []any {

	if value == nil {
		return nil
	}

	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		result := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = v.Index(i).Interface()
		}
		return result
	default:
		return []any{value}
	}
}

// ========== 新的流畅链式 API ==========

// Or 组合多个 WHERE 条件为 OR
// 接受多个 *Query 类型的参数
// 内部使用 AND 组合多个查询条件作为一个整体
//
// 示例:
//   - Or(Eq("name", "John"), Eq("name", "Jane"))
func (w *where[Q]) Or(querys ...Q) Q {

	if len(querys) == 0 {
		return w.Parent
	}

	// 提取所有 where 表达式，内部使用 AND 组合
	exprs := make([]clause.Expression, 0, len(querys))
	for _, query := range querys {

		if err := query.getError(); err != nil {
			w.Parent.setError(err)
			return w.Parent
		}

		qs := query.WhereExpr()
		if len(qs.Exprs) > 0 {
			var expr clause.Expression
			if len(qs.Exprs) == 1 {
				expr = qs.Exprs[0]
			} else {
				// 内部使用 AND 组合多个条件
				expr = clause.And(qs.Exprs...)
			}
			exprs = append(exprs, expr)
		}
	}

	if len(exprs) > 0 {
		// 外部使用 OR 组合多个表达式
		w.Value.Merge(clause.Where{Exprs: []clause.Expression{clause.Or(clause.And(exprs...))}})
	}

	return w.Parent
}

// And 组合多个 WHERE 条件为 AND
// 接受多个 *Query 类型的参数
//
// 示例:
//   - And(Eq("name", "John"), Gte("age", 18))
func (w *where[Q]) And(querys ...Q) Q {

	if len(querys) == 0 {
		return w.Parent
	}

	// 提取所有 where 表达式
	exprs := make([]clause.Expression, 0)
	for _, query := range querys {

		if err := query.getError(); err != nil {
			w.Parent.setError(err)
			return w.Parent
		}

		qs := query.WhereExpr()
		if len(qs.Exprs) > 0 {
			var expr clause.Expression
			if len(qs.Exprs) == 1 {
				expr = qs.Exprs[0]
			} else {
				expr = clause.And(qs.Exprs...)
			}
			exprs = append(exprs, expr)
		}
	}

	if len(exprs) > 0 {
		w.Value.Merge(clause.Where{Exprs: []clause.Expression{clause.And(exprs...)}})
	}

	return w.Parent
}

// Not 对 WHERE 条件取反
// 接受 *Query 类型的参数
//
// 示例:
//   - Not(Eq("city", "London"))
func (w *where[Q]) Not(query Q) Q {

	if err := query.getError(); err != nil {
		w.Parent.setError(err)
		return w.Parent
	}

	qs := query.WhereExpr()
	if len(qs.Exprs) > 0 {
		w.Value.Merge(clause.Where{Exprs: []clause.Expression{clause.Not(qs.Exprs...)}})
	}

	return w.Parent
}

// Eq 添加等于条件
func (w *where[Q]) Eq(column string, value any) Q {
	w.Value.Merge(clause.Where{Exprs: []clause.Expression{clause.Eq{Column: column, Value: value}}})
	return w.Parent
}

// Neq 添加不等于条件
func (w *where[Q]) Neq(column string, value any) Q {
	w.Value.Merge(clause.Where{Exprs: []clause.Expression{clause.Neq{Column: column, Value: value}}})
	return w.Parent
}

// Gt 添加大于条件
func (w *where[Q]) Gt(column string, value any) Q {
	w.Value.Merge(clause.Where{Exprs: []clause.Expression{clause.Gt{Column: column, Value: value}}})
	return w.Parent
}

// Gte 添加大于等于条件
func (w *where[Q]) Gte(column string, value any) Q {
	w.Value.Merge(clause.Where{Exprs: []clause.Expression{clause.Gte{Column: column, Value: value}}})
	return w.Parent
}

// Lt 添加小于条件
func (w *where[Q]) Lt(column string, value any) Q {
	w.Value.Merge(clause.Where{Exprs: []clause.Expression{clause.Lt{Column: column, Value: value}}})
	return w.Parent
}

// Lte 添加小于等于条件
func (w *where[Q]) Lte(column string, value any) Q {
	w.Value.Merge(clause.Where{Exprs: []clause.Expression{clause.Lte{Column: column, Value: value}}})
	return w.Parent
}

// Like 添加 LIKE 条件
func (w *where[Q]) Like(column string, value any) Q {
	w.Value.Merge(clause.Where{Exprs: []clause.Expression{clause.Like{Column: column, Value: value}}})
	return w.Parent
}

// In 添加 IN 条件
// 支持传入展开值或切片/数组，内部会自动转换为 IN 条件
//
// 示例：
//   - In("city", "London", "Paris", "Berlin")
//   - In("id", []int{1, 2, 3})
func (w *where[Q]) In(column string, values ...any) Q {
	if len(values) == 1 {
		w.Value.Merge(clause.Where{Exprs: []clause.Expression{clause.IN{Column: column, Values: toAnySlice(values[0])}}})
	} else if len(values) > 1 {
		w.Value.Merge(clause.Where{Exprs: []clause.Expression{clause.IN{Column: column, Values: values}}})
	}
	return w.Parent
}

// ========== 全局逻辑组合函数 ==========

// Or 组合多个 WHERE 条件为 OR
func Or(querys ...*Query) *Query {
	return newQuery("").Or(querys...)
}

// And 组合多个 WHERE 条件为 AND
func And(querys ...*Query) *Query {
	return newQuery("").And(querys...)
}

// Not 对 WHERE 条件取反
func Not(query *Query) *Query {
	return newQuery("").Not(query)
}

// ========== 全局比较操作符函数 ==========

// Eq 创建等于条件
func Eq(column string, value any) *Query {
	return newQuery("").Eq(column, value)
}

// Neq 创建不等于条件
func Neq(column string, value any) *Query {
	return newQuery("").Neq(column, value)
}

// Gt 创建大于条件
func Gt(column string, value any) *Query {
	return newQuery("").Gt(column, value)
}

// Gte 创建大于等于条件
func Gte(column string, value any) *Query {
	return newQuery("").Gte(column, value)
}

// Lt 创建小于条件
func Lt(column string, value any) *Query {
	return newQuery("").Lt(column, value)
}

// Lte 创建小于等于条件
func Lte(column string, value any) *Query {
	return newQuery("").Lte(column, value)
}

// Like 创建 LIKE 条件
func Like(column string, value any) *Query {
	return newQuery("").Like(column, value)
}

// In 添加 IN 条件
// 支持传入展开值或切片/数组，内部会自动转换为 IN 条件
//
// 示例：
//   - In("city", "London", "Paris", "Berlin")
//   - In("id", []int{1, 2, 3})
func In(column string, values ...any) *Query {
	return newQuery("").In(column, values...)
}
