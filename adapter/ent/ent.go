package ent

import (
	"errors"

	"entgo.io/ent/dialect/sql"
	"github.com/epkgs/query/clause"
)

// Where 将 clause.Where 转换为 ent 的 Where 函数
func Where(where clause.Where) func(s *sql.Selector) {
	return func(s *sql.Selector) {
		if len(where.Exprs) == 0 {
			return
		}

		// 将 query/clause.Where 转换为 ent 的条件
		pred, err := convertToEntWhere(where)
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
func convertToEntWhere(where clause.Where) (*sql.Predicate, error) {
	if len(where.Exprs) == 0 {
		return nil, nil
	}

	var preds []*sql.Predicate
	for _, expr := range where.Exprs {
		pred, err := convertToEntPredicate(expr)
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
func convertToEntPredicate(expr clause.Expression) (*sql.Predicate, error) {
	switch e := expr.(type) {
	case clause.Eq:
		return sql.EQ(e.Column, e.Value), nil
	case clause.Neq:
		return sql.NEQ(e.Column, e.Value), nil
	case clause.Gt:
		return sql.GT(e.Column, e.Value), nil
	case clause.Gte:
		return sql.GTE(e.Column, e.Value), nil
	case clause.Lt:
		return sql.LT(e.Column, e.Value), nil
	case clause.Lte:
		return sql.LTE(e.Column, e.Value), nil
	case clause.Like:
		// 将 interface{} 转换为 string
		if likeValue, ok := e.Value.(string); ok {
			return sql.Like(e.Column, likeValue), nil
		}
		return nil, errors.New("like value must be string")
	case clause.IN:
		return sql.In(e.Column, e.Values...), nil
	case clause.AndExpr:
		// 对于 AND 表达式，递归处理所有子表达式
		if len(e.Exprs) == 0 {
			return nil, nil
		}
		// 初始化 AND 条件切片
		var andPreds []*sql.Predicate
		for _, subExpr := range e.Exprs {
			pred, err := convertToEntPredicate(subExpr)
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
			pred, err := convertToEntPredicate(subExpr)
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
			pred, err := convertToEntPredicate(subExpr)
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
func OrderBy(orders clause.OrderBys) func(s *sql.Selector) {
	return func(s *sql.Selector) {
		if len(orders) == 0 {
			return
		}

		for _, order := range orders {
			if order.Desc {
				s.OrderBy(sql.Desc(order.Column))
			} else {
				s.OrderBy(sql.Asc(order.Column))
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
func Query(where clause.Where, orders clause.OrderBys, pagination clause.Pagination) func(s *sql.Selector) {
	return func(s *sql.Selector) {
		Where(where)(s)
		OrderBy(orders)(s)
		Pagination(pagination)(s)
	}
}
