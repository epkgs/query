package ent

import (
	"errors"

	"entgo.io/ent/dialect/sql"
	"github.com/epkgs/query/clause"
)

type options struct {
	mapper MapperFunc
}

type Option func(*options)

// MapperFunc 定义了查询列名到 ENT 列名的映射函数类型
// 输入参数 queryColumn 是查询条件中使用的列名
// 返回值 entColumn 是数据库表中的实际列名
type MapperFunc func(queryColumn string) (entColumn string)

// WithMapper 创建一个配置选项，用于设置查询列名到数据库实际列名的映射函数
// mapper 参数是一个 MapperFunc 类型的函数，用于自定义列名映射逻辑
// 该选项用于在将查询条件转换为 ENT 条件时，通过自定义函数实现灵活的列名映射
// 例如：可以实现驼峰命名转下划线命名、添加前缀/后缀等复杂映射逻辑
func WithMapper(mapper MapperFunc) Option {
	// 返回一个 Option 函数类型，该函数接收一个 options 指针并修改其 mapper 字段
	return func(o *options) {
		o.mapper = mapper
	}
}

func (o *options) Column(column string) string {
	if o.mapper != nil {
		return o.mapper(column)
	}

	return column
}

// Where 将 clause.Where 转换为 ent 的 Where 函数
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

	var preds []*sql.Predicate
	for _, expr := range where.Exprs {
		pred, err := convertToEntPredicate(expr, opt)
		if err != nil {
			return nil, err
		}
		if pred != nil {
			preds = append(preds, pred)
		}
	}

	if len(preds) == 1 {
		return preds[0], nil
	} else if len(preds) > 1 {
		return sql.And(preds...), nil
	}

	return nil, nil
}

// convertToEntPredicate 将 query/clause.Expression 转换为 *sql.Predicate
func convertToEntPredicate(expr clause.Expression, opt *options) (*sql.Predicate, error) {
	switch e := expr.(type) {
	case clause.Eq:
		return sql.EQ(opt.Column(e.Column), e.Value), nil
	case clause.Neq:
		return sql.NEQ(opt.Column(e.Column), e.Value), nil
	case clause.Gt:
		return sql.GT(opt.Column(e.Column), e.Value), nil
	case clause.Gte:
		return sql.GTE(opt.Column(e.Column), e.Value), nil
	case clause.Lt:
		return sql.LT(opt.Column(e.Column), e.Value), nil
	case clause.Lte:
		return sql.LTE(opt.Column(e.Column), e.Value), nil
	case clause.Like:
		// 将 interface{} 转换为 string
		if likeValue, ok := e.Value.(string); ok {
			return sql.Like(opt.Column(e.Column), likeValue), nil
		}
		return nil, errors.New("like value must be string")
	case clause.IN:
		return sql.In(opt.Column(e.Column), e.Values...), nil
	case clause.AndExpr:
		// 对于 AND 表达式，递归处理所有子表达式
		if len(e.Exprs) == 0 {
			return nil, nil
		}
		// 初始化 AND 条件切片
		var andPreds []*sql.Predicate
		for _, subExpr := range e.Exprs {
			pred, err := convertToEntPredicate(subExpr, opt)
			if err != nil {
				return nil, err
			}
			if pred != nil {
				andPreds = append(andPreds, pred)
			}
		}
		// 如果只有一个条件，直接返回
		if len(andPreds) == 1 {
			return andPreds[0], nil
		}
		// 否则返回 AND 连接的条件
		return sql.And(andPreds...), nil
	case clause.OrExpr:
		// 对于 OR 表达式，递归处理所有子表达式
		if len(e.Exprs) == 0 {
			return nil, nil
		}
		// 初始化 OR 条件切片
		var orPreds []*sql.Predicate
		for _, subExpr := range e.Exprs {
			pred, err := convertToEntPredicate(subExpr, opt)
			if err != nil {
				return nil, err
			}
			if pred != nil {
				orPreds = append(orPreds, pred)
			}
		}
		// 如果只有一个条件，直接返回
		if len(orPreds) == 1 {
			return orPreds[0], nil
		}
		// 否则返回 OR 连接的条件
		return sql.Or(orPreds...), nil
	case clause.NotExpr:
		// 对于 NOT 表达式，递归处理所有子表达式
		if len(e.Exprs) == 0 {
			return nil, nil
		}
		// 初始化 NOT 条件切片
		var notPreds []*sql.Predicate
		for _, subExpr := range e.Exprs {
			pred, err := convertToEntPredicate(subExpr, opt)
			if err != nil {
				return nil, err
			}
			if pred != nil {
				notPreds = append(notPreds, pred)
			}
		}
		// 如果只有一个条件，直接返回 NOT 条件
		if len(notPreds) == 1 {
			return sql.Not(notPreds[0]), nil
		}
		// 否则返回 NOT AND 连接的条件
		return sql.Not(sql.And(notPreds...)), nil
	default:
		return nil, nil
	}
}

// OrderBy 将 clause.OrderBys 转换为 ent 的 OrderBy 函数
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
			if order.Desc {
				s.OrderBy(sql.Desc(opt.Column(order.Column)))
			} else {
				s.OrderBy(sql.Asc(opt.Column(order.Column)))
			}
		}
	}
}

// Pagination 将 clause.Pagination 转换为 ent 的 Pagination 函数
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

// Query 将多个查询组件转换为 ent 的查询函数
func Query(where clause.Where, orders clause.OrderBys, pagination clause.Pagination, opts ...Option) func(s *sql.Selector) {
	return func(s *sql.Selector) {
		Where(where, opts...)(s)
		OrderBy(orders, opts...)(s)
		Pagination(pagination)(s)
	}
}
