// Package ent 提供了将 query/clause 查询组件转换为 Ent ORM 查询条件的适配器。
//
// 该适配器将 clause.Where、clause.OrderBys、clause.Pagination 转换为
// Ent 的 sql.Selector 修改函数（func(*sql.Selector)），
// 可直接用于 ent/client 的 Modify 方法或 sql.Selector 操作。
//
// 与 AIP 配合使用的典型流程：
//
//	import (
//	    aip "github.com/epkgs/query/adapter/aip"
//	    entadapter "github.com/epkgs/query/adapter/ent"
//	)
//
//	// 1. 解析 AIP 过滤和排序
//	filter, _ := filtering.ParseFilter(request, declarations)
//	whereClause, _ := aip.FromFilter(filter)
//	orderBys := aip.FromOrderBy(parsedOrderBy)
//
//	// 2. 使用 ExprHandler 映射字段名
//	handler := func(expr clause.Expression) clause.Expression {
//	    if eq, ok := expr.(clause.Eq); ok {
//	        eq.Column = fieldMappings[eq.Column]
//	    }
//	    return expr
//	}
//
//	// 3. 应用到 Ent 查询
//	client.User.Query().Modify(entadapter.Query(
//	    whereClause, orderBys, clause.Pagination{},
//	    entadapter.WithExprHandler(handler),
//	)).All(ctx)
package ent

import (
	"errors"

	"entgo.io/ent/dialect/sql"
	"github.com/epkgs/query/clause"
)

type options struct {
	exprHandler  ExprHandler
	orderHandler OrderHandler
}

type Option func(*options)

// ExprHandler 表达式处理器函数类型。
// 在转换为 Ent Predicate 前对 clause.Expression 进行预处理，
// 通常用于将 clause 的字段名映射为 ent schema 定义的列名。
//
// 示例（字段名映射）：
//
//	handler := func(expr clause.Expression) clause.Expression {
//	    switch e := expr.(type) {
//	    case clause.Eq:
//	        e.Column = fieldMappings[e.Column]
//	        return e
//	    case clause.Like:
//	        e.Column = fieldMappings[e.Column]
//	        return e
//	    }
//	    return expr
//	}
type ExprHandler func(expr clause.Expression) clause.Expression

// WithExprHandler 设置表达式处理器。
// 在 Where/Query 转换过程中，每个 clause.Expression 都会经过该处理器，
// 常用于将 API 暴露的字段名映射为 ent schema 的实际数据库列名。
func WithExprHandler(handler ExprHandler) Option {
	return func(o *options) {
		o.exprHandler = handler
	}
}

// OrderHandler 排序处理器函数类型。
// 在转换为 Ent 排序条件前对 clause.OrderBy 进行预处理，
// 通常用于将排序字段名映射为 ent schema 的实际列名。
type OrderHandler func(expr clause.OrderBy) clause.OrderBy

// WithOrderByHandler 设置排序处理器。
// 在 OrderBy/Query 转换过程中，每个 clause.OrderBy 都会经过该处理器。
func WithOrderByHandler(handler OrderHandler) Option {
	return func(o *options) {
		o.orderHandler = handler
	}
}

// Where 将 clause.Where 转换为 Ent 的 sql.Selector 修改函数。
// 支持通过 Option 设置 ExprHandler/OrderHandler 进行字段映射。
// 返回的函数可直接传入 client.Query().Modify() 或 sql.Selector 操作。
func Where(where clause.Where, opts ...Option) func(s *sql.Selector) {

	opt := &options{}
	for _, o := range opts {
		o(opt)
	}

	return func(s *sql.Selector) {
		if len(where.Exprs) == 0 {
			return
		}

		// 将 query/clause.Where 转换为 ent 的条件
		pred, err := convertToEntWhere(where, opt)
		if err != nil {
			s.Builder.AddError(err)
			return
		}

		if pred != nil {
			s.Where(pred)
		}
	}
}

// convertToEntWhere 将 query/clause.Where 转换为 ent 的条件
func convertToEntWhere(where clause.Where, opt *options) (*sql.Predicate, error) {
	if len(where.Exprs) == 0 {
		return nil, nil
	}

	var err error
	var pred *sql.Predicate
	for _, expr := range where.Exprs {
		pred, err = convertToEntPredicate(pred, expr, opt)
		if err != nil {
			return nil, err
		}
	}

	return pred, nil
}

func sqlAnd(pred1, pred2 *sql.Predicate) *sql.Predicate {
	if pred1 != nil {
		return sql.And(pred1, pred2)
	} else {
		return pred2
	}
}

func sqlOr(pred1, pred2 *sql.Predicate) *sql.Predicate {
	if pred1 != nil {
		return sql.Or(pred1, pred2)
	} else {
		return pred2
	}
}

// convertToEntPredicate 将 query/clause.Expression 转换为 *sql.Predicate
func convertToEntPredicate(pre *sql.Predicate, expr clause.Expression, opt *options) (*sql.Predicate, error) {

	if opt.exprHandler != nil {
		expr = opt.exprHandler(expr) // 调用转换器
	}

	if expr == nil {
		return pre, nil
	}

	switch e := expr.(type) {
	case clause.Eq:
		return sqlAnd(pre, sql.EQ(e.Column, e.Value)), nil
	case clause.Neq:
		return sqlAnd(pre, sql.NEQ(e.Column, e.Value)), nil
	case clause.Gt:
		return sqlAnd(pre, sql.GT(e.Column, e.Value)), nil
	case clause.Gte:
		return sqlAnd(pre, sql.GTE(e.Column, e.Value)), nil
	case clause.Lt:
		return sqlAnd(pre, sql.LT(e.Column, e.Value)), nil
	case clause.Lte:
		return sqlAnd(pre, sql.LTE(e.Column, e.Value)), nil
	case clause.Like:
		// 将 interface{} 转换为 string
		if likeValue, ok := e.Value.(string); ok {
			return sqlAnd(pre, sql.Like(e.Column, likeValue)), nil
		}
		return nil, errors.New("like value must be string")
	case clause.IN:
		return sqlAnd(pre, sql.In(e.Column, e.Values...)), nil
	case clause.AndExpr:
		// 对于 AND 表达式，递归处理所有子表达式
		if len(e.Exprs) == 0 {
			return pre, nil
		}
		// 初始化 AND 条件
		var subPred *sql.Predicate
		var err error
		for _, subExpr := range e.Exprs {
			subPred, err = convertToEntPredicate(subPred, subExpr, opt)
			if err != nil {
				return nil, err
			}
		}

		if pre != nil {
			return sql.And(pre, subPred), nil
		}
		return subPred, nil
	case clause.OrExpr:
		// 对于 OR 表达式，递归处理所有子表达式
		if len(e.Exprs) == 0 {
			return nil, nil
		}
		// 初始化 OR 条件
		var subPred *sql.Predicate
		var err error
		for _, subExpr := range e.Exprs {
			subPred, err = convertToEntPredicate(subPred, subExpr, opt)
			if err != nil {
				return nil, err
			}
		}

		return sqlOr(pre, subPred), nil

	case clause.NotExpr:
		// 对于 NOT 表达式，递归处理所有子表达式
		if len(e.Exprs) == 0 {
			return nil, nil
		}

		// 初始化 NOT 条件
		var subPred *sql.Predicate
		var err error
		for _, subExpr := range e.Exprs {
			subPred, err = convertToEntPredicate(subPred, subExpr, opt)
			if err != nil {
				return nil, err
			}
		}

		return sqlAnd(pre, sql.Not(subPred)), nil
	default:
		return nil, nil
	}
}

// OrderBy 将 clause.OrderBys 转换为 Ent 的排序函数。
// 支持通过 Option 设置 OrderHandler 进行字段映射。
func OrderBy(orders clause.OrderBys, opts ...Option) func(s *sql.Selector) {
	opt := &options{}
	for _, o := range opts {
		o(opt)
	}

	return func(s *sql.Selector) {
		if len(orders) == 0 {
			return
		}

		for _, order := range orders {

			if opt.orderHandler != nil {
				*order = opt.orderHandler(*order)
			}

			if order.Column == "" {
				continue
			}

			if order.Desc {
				s.OrderBy(sql.Desc(order.Column))
			} else {
				s.OrderBy(sql.Asc(order.Column))
			}
		}
	}
}

// Pagination 将 clause.Pagination 转换为 Ent 的 LIMIT/OFFSET 设置函数。
func Pagination(pagination clause.Pagination) func(s *sql.Selector) {
	return func(s *sql.Selector) {

		if pagination.Limit != nil {
			s.Limit(*pagination.Limit)
		}

		if pagination.Offset > 0 {
			s.Offset(pagination.Offset)
		}
	}
}

// Query 将 WHERE、ORDER BY 和分页三个查询组件一次性转换为 Ent sql.Selector 修改函数。
// 这是 Where、OrderBy、Pagination 三个函数的便捷组合。
// 支持通过 Option 设置 ExprHandler/OrderHandler 进行字段映射。
func Query(where clause.Where, orders clause.OrderBys, pagination clause.Pagination, opts ...Option) func(s *sql.Selector) {
	return func(s *sql.Selector) {
		Where(where, opts...)(s)
		OrderBy(orders, opts...)(s)
		Pagination(pagination)(s)
	}
}
