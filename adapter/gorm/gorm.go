package gorm

import (
	"github.com/epkgs/query/clause"
	"gorm.io/gorm"
	gormClause "gorm.io/gorm/clause"
)

type options struct {
	exprHandler  ExprHandler
	orderHandler OrderHandler
}

type Option func(*options)

// ExprHandler 表达式处理器函数类型，用于在转换为GORM表达式前预处理clause.Expression
// expr: 原始的查询表达式
// 返回值: 预处理后的查询表达式
type ExprHandler func(expr clause.Expression) clause.Expression

func WithExprHandler(handler ExprHandler) Option {
	return func(o *options) {
		o.exprHandler = handler
	}
}

// OrderHandler 排序处理器函数类型，用于在转换为GORM排序表达式前预处理clause.OrderBy
// expr: 原始的排序表达式
// 返回值: 预处理后的排序表达式
type OrderHandler func(expr clause.OrderBy) clause.OrderBy

func WithOrderByHandler(handler OrderHandler) Option {
	return func(o *options) {
		o.orderHandler = handler
	}
}

// Where 将 clause.Where 转换为 gorm scope
func Where(where clause.Where, opts ...Option) func(db *gorm.DB) *gorm.DB {

	opt := &options{}
	for _, o := range opts {
		o(opt)
	}

	return func(db *gorm.DB) *gorm.DB {
		if len(where.Exprs) == 0 {
			return db
		}

		// 将 query/clause.Where 转换为 gorm/clause.Where
		gormWhere := convertWhere(where, opt)

		// 将 gorm/clause.Where 给到 gorm.DB 的 Where 函数
		return db.Where(gormWhere)
	}
}

// convertWhere 将 query/clause.Where 转换为 gorm/clause.Where
func convertWhere(where clause.Where, opt *options) gormClause.Where {
	gormExprs := make([]gormClause.Expression, 0, len(where.Exprs))

	for _, expr := range where.Exprs {
		gormExpr := convertExpr(expr, opt)
		if gormExpr != nil {
			gormExprs = append(gormExprs, gormExpr)
		}
	}

	return gormClause.Where{Exprs: gormExprs}
}

// convertExpr 将 query/clause.Expression 转换为 gorm/clause.Expression
func convertExpr(expr clause.Expression, opt *options) gormClause.Expression {

	if opt.exprHandler != nil {
		expr = opt.exprHandler(expr) // 调用转换器
	}

	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case clause.Eq:
		return gormClause.Eq{Column: gormClause.Column{Name: e.Column}, Value: e.Value}
	case clause.Neq:
		return gormClause.Neq{Column: gormClause.Column{Name: e.Column}, Value: e.Value}
	case clause.Gt:
		return gormClause.Gt{Column: gormClause.Column{Name: e.Column}, Value: e.Value}
	case clause.Gte:
		return gormClause.Gte{Column: gormClause.Column{Name: e.Column}, Value: e.Value}
	case clause.Lt:
		return gormClause.Lt{Column: gormClause.Column{Name: e.Column}, Value: e.Value}
	case clause.Lte:
		return gormClause.Lte{Column: gormClause.Column{Name: e.Column}, Value: e.Value}
	case clause.Like:
		return gormClause.Like{Column: gormClause.Column{Name: e.Column}, Value: e.Value}
	case clause.IN:
		return gormClause.IN{Column: gormClause.Column{Name: e.Column}, Values: e.Values}
	case clause.AndExpr:
		var gormExprs []gormClause.Expression
		for _, subExpr := range e.Exprs {
			gormExpr := convertExpr(subExpr, opt)
			if gormExpr != nil {
				gormExprs = append(gormExprs, gormExpr)
			}
		}
		if len(gormExprs) == 0 {
			return nil
		}
		return gormClause.And(gormExprs...)
	case clause.OrExpr:
		var gormExprs []gormClause.Expression
		for _, subExpr := range e.Exprs {
			gormExpr := convertExpr(subExpr, opt)
			if gormExpr != nil {
				gormExprs = append(gormExprs, gormExpr)
			}
		}
		if len(gormExprs) == 0 {
			return nil
		}
		return gormClause.Or(gormExprs...)
	case clause.NotExpr:
		var gormExprs []gormClause.Expression
		for _, subExpr := range e.Exprs {
			gormExpr := convertExpr(subExpr, opt)
			if gormExpr != nil {
				gormExprs = append(gormExprs, gormExpr)
			}
		}
		if len(gormExprs) == 0 {
			return nil
		}
		return gormClause.Not(gormExprs...)
	default:
		return nil
	}
}

// OrderBy 将 clause.OrderBys 转换为 gorm scope
func OrderBy(orders clause.OrderBys, opts ...Option) func(db *gorm.DB) *gorm.DB {
	opt := &options{}
	for _, o := range opts {
		o(opt)
	}

	return func(db *gorm.DB) *gorm.DB {
		if len(orders) == 0 {
			return db
		}

		gOrderByCols := []gormClause.OrderByColumn{}

		for _, order := range orders {
			if opt.orderHandler != nil {
				order = opt.orderHandler(order) // 调用转换器
			}

			if order.Column == "" {
				continue
			}
			gOrderByCols = append(gOrderByCols, gormClause.OrderByColumn{
				Column: gormClause.Column{Name: order.Column},
				Desc:   order.Desc,
			})
		}

		db.Order(gormClause.OrderBy{Columns: gOrderByCols})

		return db
	}
}

// Pagination 将 clause.Pagination 转换为 gorm scope
func Pagination(pagination clause.Pagination) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		db.Statement.AddClause(gormClause.Limit{
			Limit:  pagination.Limit,
			Offset: pagination.Offset,
		})

		return db
	}
}

// Query 将多个查询组件转换为 gorm scope
func Query(where clause.Where, orders clause.OrderBys, pagination clause.Pagination, opts ...Option) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db = Where(where, opts...)(db)
		db = OrderBy(orders, opts...)(db)
		db = Pagination(pagination)(db)
		return db
	}
}
