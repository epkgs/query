// Package gorm 提供了将 query/clause 查询组件转换为 GORM Scope 函数的适配器。
//
// 该适配器将 clause.Where、clause.OrderBys、clause.Pagination 分别转换为
// gorm.DB 的 Scope 函数（func(*gorm.DB) *gorm.DB），
// 可直接用于 db.Scopes() 方法中。
//
// 使用方式：
//
//	q := query.Table("users").Where("age", ">", 18).Select("id", "name")
//	db.Scopes(
//	    gormadapter.Where(q.WhereExpr()),
//	    gormadapter.OrderBy(q.OrderByExpr()),
//	    gormadapter.Pagination(q.PaginationExpr()),
//	).Find(&users)
//
// 组合使用：
//
//	db.Scopes(gormadapter.Query(q.WhereExpr(), q.OrderByExpr(), q.PaginationExpr())).Find(&users)
//
// 与 AIP 配合使用：
//
//	import (
//	    aip "github.com/epkgs/query/adapter/aip"
//	    gorm "github.com/epkgs/query/adapter/gorm"
//	)
//	filter, _ := filtering.ParseFilter(request, declarations)
//	whereClause, _ := aip.FromFilter(filter)
//	orderBys := aip.FromOrderBy(parsedOrderBy)
//	db.Scopes(gorm.Query(whereClause, orderBys, clause.Pagination{})).Find(&users)
package gorm

import (
	"github.com/epkgs/query/clause"
	"gorm.io/gorm"
	gormClause "gorm.io/gorm/clause"
)

// Where 将 clause.Where 转换为 GORM Scope 函数。
// 返回的函数可直接传入 db.Scopes() 或 db.Where() 使用。
// 如果 where 没有表达式，返回空操作的 Scope。
func Where(where clause.Where) func(db *gorm.DB) *gorm.DB {

	return func(db *gorm.DB) *gorm.DB {
		if len(where.Exprs) == 0 {
			return db
		}

		// 将 query/clause.Where 转换为 gorm/clause.Where
		gormWhere := convertWhere(where)

		if len(gormWhere.Exprs) == 0 {
			return db
		}

		// 将 gorm/clause.Where 给到 gorm.DB 的 Where 函数
		return db.Where(gormWhere)
	}
}

// convertWhere 将 query/clause.Where 转换为 gorm/clause.Where
func convertWhere(where clause.Where) gormClause.Where {
	gormExprs := make([]gormClause.Expression, 0, len(where.Exprs))

	for _, expr := range where.Exprs {
		gormExpr := convertExpr(expr)
		if gormExpr != nil {
			gormExprs = append(gormExprs, gormExpr)
		}
	}

	return gormClause.Where{Exprs: gormExprs}
}

// convertExpr 将 query/clause.Expression 转换为 gorm/clause.Expression
func convertExpr(expr clause.Expression) gormClause.Expression {

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
			gormExpr := convertExpr(subExpr)
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
			gormExpr := convertExpr(subExpr)
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
			gormExpr := convertExpr(subExpr)
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

// OrderBy 将 clause.OrderBys 转换为 GORM Scope 函数，用于设置排序条件。
func OrderBy(orders clause.OrderBys) func(db *gorm.DB) *gorm.DB {

	return func(db *gorm.DB) *gorm.DB {
		if len(orders) == 0 {
			return db
		}

		gOrderByCols := []gormClause.OrderByColumn{}

		for _, order := range orders {

			if order.Column == "" {
				continue
			}
			gOrderByCols = append(gOrderByCols, gormClause.OrderByColumn{
				Column: gormClause.Column{Name: order.Column},
				Desc:   order.Desc,
			})
		}

		if len(gOrderByCols) > 0 {
			db.Order(gormClause.OrderBy{Columns: gOrderByCols})
		}

		return db
	}
}

// Pagination 将 clause.Pagination 转换为 GORM Scope 函数，用于设置 LIMIT 和 OFFSET。
func Pagination(pagination clause.Pagination) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		db.Statement.AddClause(gormClause.Limit{
			Limit:  pagination.Limit,
			Offset: pagination.Offset,
		})

		return db
	}
}

// Query 将 WHERE、ORDER BY 和分页三个查询组件一次性转换为 GORM Scope 函数。
// 这是 Where、OrderBy、Pagination 三个函数的便捷组合。
func Query(where clause.Where, orders clause.OrderBys, pagination clause.Pagination) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db = Where(where)(db)
		db = OrderBy(orders)(db)
		db = Pagination(pagination)(db)
		return db
	}
}
