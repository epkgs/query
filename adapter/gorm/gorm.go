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
//	    gormadapter.WhereScope(q.WhereExpr()),
//	    gormadapter.OrderByScope(q.OrderByExpr()),
//	    gormadapter.PaginationScope(q.PaginationExpr()),
//	).Find(&users)
//
// 组合使用：
//
//	db.Scopes(gormadapter.QueryScope(q.WhereExpr(), q.OrderByExpr(), q.PaginationExpr())).Find(&users)
//
// 与 AIP 配合使用：
//
//	import (
//	    aip "github.com/epkgs/query/adapter/aip"
//	    gormadapter "github.com/epkgs/query/adapter/gorm"
//	)
//	filter, _ := filtering.ParseFilter(request, declarations)
//	whereClause, _ := aip.FromFilter(filter)
//	orderBys := aip.FromOrderBy(parsedOrderBy)
//	db.Scopes(gormadapter.QueryScope(whereClause, orderBys, clause.Pagination{})).Find(&users)
package gorm

import (
	"github.com/epkgs/query/clause"
	"gorm.io/gorm"
	gormClause "gorm.io/gorm/clause"
)

// WhereConverter 将 Expression 转换为 GORM 的 Expression；若转换成功 converted 为 true，否则由默认逻辑处理。
type WhereConverter func(e clause.Expression, c *clause.Condition) (gormExpr gormClause.Expression, converted bool)

// OrderByConverter 将 OrderBy 转换为 GORM 的 OrderByColumn；若转换成功 converted 为 true，否则由默认逻辑处理。
type OrderByConverter func(o clause.OrderBy) (gormOrder gormClause.OrderByColumn, converted bool)

func WhereExpr(where clause.Where, convs ...WhereConverter) gormClause.Expression {
	if len(where.Exprs) == 0 {
		return nil
	}

	// 将 query/clause.Where 转换为 gorm/clause.Where
	gormExprs := convertWhere(where.Exprs, convs...)
	if len(gormExprs) == 0 {
		return nil
	}

	return gormClause.Where{Exprs: gormExprs}
}

// WhereScope 将 clause.Where 转换为 GORM Scope 函数。
// 返回的函数可直接传入 db.Scopes() 或 db.Where() 使用。
// 如果 where 没有表达式，返回空操作的 Scope。
func WhereScope(where clause.Where, convs ...WhereConverter) func(db *gorm.DB) *gorm.DB {

	return func(db *gorm.DB) *gorm.DB {

		gormWhere := WhereExpr(where, convs...)

		if gormWhere == nil {
			return db
		}

		return db.Where(gormWhere)
	}
}

// convertWhere 将 query/clause.Where 转换为 gorm/clause.Where
func convertWhere(exprs []clause.Expression, convs ...WhereConverter) []gormClause.Expression {
	gormExprs := make([]gormClause.Expression, 0, len(exprs))

	for _, expr := range exprs {
		var gormExpr gormClause.Expression
		var converted bool
		c, _ := clause.AsCondition(expr)
		for _, conv := range convs {
			gormExpr, converted = conv(expr, c)
			if converted {
				break
			}
		}

		if !converted {
			gormExpr = convertExpr(expr)
		}

		if gormExpr != nil {
			gormExprs = append(gormExprs, gormExpr)
		}
	}

	return gormExprs
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

func OrderByExpr(orders clause.OrderBys, convs ...OrderByConverter) gormClause.OrderBy {
	gOrderByCols := []gormClause.OrderByColumn{}

	for _, order := range orders {

		if order == nil {
			continue
		}

		if order.Column == "" {
			continue
		}

		var gcol gormClause.OrderByColumn
		var converted bool

		for _, conv := range convs {
			gcol, converted = conv(*order)
			if converted {
				break
			}
		}

		if !converted {
			gcol = gormClause.OrderByColumn{
				Column: gormClause.Column{Name: order.Column},
				Desc:   order.Desc,
			}
		}

		gOrderByCols = append(gOrderByCols, gcol)
	}

	return gormClause.OrderBy{Columns: gOrderByCols}
}

// OrderByScope 将 clause.OrderBys 转换为 GORM Scope 函数，用于设置排序条件。
func OrderByScope(orders clause.OrderBys, convs ...OrderByConverter) func(db *gorm.DB) *gorm.DB {

	return func(db *gorm.DB) *gorm.DB {
		gormOrderBy := OrderByExpr(orders, convs...)

		if len(gormOrderBy.Columns) > 0 {
			return db.Order(gormOrderBy)
		}

		return db
	}
}

// PaginationScope 将 clause.Pagination 转换为 GORM Scope 函数，用于设置 LIMIT 和 OFFSET。
func PaginationScope(pagination clause.Pagination) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if pagination.Limit == nil && pagination.Offset == 0 {
			return db
		}

		db.Statement.AddClause(gormClause.Limit{
			Limit:  pagination.Limit,
			Offset: pagination.Offset,
		})

		return db
	}
}

// QueryScope 将 WHERE、ORDER BY 和分页三个查询组件一次性转换为 GORM Scope 函数。
// 这是 WhereScope、OrderByScope、PaginationScope 三个函数的便捷组合。
func QueryScope(where clause.Where, orders clause.OrderBys, pagination clause.Pagination) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db = WhereScope(where)(db)
		db = OrderByScope(orders)(db)
		db = PaginationScope(pagination)(db)
		return db
	}
}
