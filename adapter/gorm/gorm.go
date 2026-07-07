// Package gorm 提供了将 query/clause 查询组件转换为 GORM 表达式和 Scope 函数的适配器。
//
// 该适配器提供两类转换：
//   - 表达式级转换：WhereExpr、OrderByExpr 将单个表达式转换为 GORM Expression，可传入自定义转换器；
//   - Scope 级转换：WhereScope、OrderByScope、PaginationScope 将查询组件转换为
//     gorm.DB 的 Scope 函数（func(*gorm.DB) *gorm.DB），可直接用于 db.Scopes() 方法中。
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
// 自定义转换：
//
//	// 自定义 WhereConverter，将 JSON 字段转换为 GORM 的 JSON 查询表达式
//	jsonConv := func(e clause.Expression, c *clause.Condition) (gormExpr gormClause.Expression, converted bool) {
//	    if c != nil && c.Column == "tags" {
//	        return gormClause.Expr{SQL: "JSON_CONTAINS(tags, ?)", Vars: []interface{}{c.Values[0]}}, true
//	    }
//	    return nil, false
//	}
//	db.Scopes(gormadapter.WhereScope(q.WhereExpr(), jsonConv)).Find(&users)
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

// WhereExpr 将单个 clause.Expression 转换为 GORM 的 Expression。
// 如果 expr 为 nil，返回 nil。
//
// 可传入 WhereConverter 对特定表达式进行自定义转换；
// 转换器按顺序执行，第一个返回 converted=true 的转换器结果即为最终结果，
// 若所有转换器均未转换，则使用默认逻辑处理。
func WhereExpr(expr clause.Expression, convs ...WhereConverter) gormClause.Expression {

	if expr == nil {
		return nil
	}

	if len(convs) > 0 {
		c, _ := clause.AsCondition(expr)
		for _, conv := range convs {
			gormExpr, converted := conv(expr, c)
			if converted {
				return gormExpr
			}
		}
	}

	return convertExpr(expr)
}

// WhereExprs 批量将 clause.Expression 列表转换为 GORM 的 Expression 列表。
// 转换过程中会跳过 nil 以及转换结果为 nil 的表达式。
//
// 可传入 WhereConverter 对特定表达式进行自定义转换，规则同 WhereExpr。
func WhereExprs(exprs []clause.Expression, convs ...WhereConverter) []gormClause.Expression {
	gormExprs := make([]gormClause.Expression, 0, len(exprs))

	for _, expr := range exprs {
		if expr == nil {
			continue
		}
		e := WhereExpr(expr, convs...)
		if e != nil {
			gormExprs = append(gormExprs, e)
		}
	}
	return gormExprs
}

// WhereScope 将 clause.Where 转换为 GORM Scope 函数。
// 返回的函数可直接传入 db.Scopes() 或 db.Where() 使用。
// 如果 where 没有表达式，返回空操作的 Scope。
//
// 可传入 WhereConverter 对特定表达式进行自定义转换。
func WhereScope(where clause.Where, convs ...WhereConverter) func(db *gorm.DB) *gorm.DB {

	return func(db *gorm.DB) *gorm.DB {
		exprs := WhereExprs(where.Exprs, convs...)

		if len(exprs) == 0 {
			return db
		}

		return db.Where(gormClause.Where{Exprs: exprs})
	}
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

// OrderByExpr 将单个 clause.OrderBy 转换为 GORM 的 OrderByColumn。
//
// 可传入 OrderByConverter 对特定排序条件进行自定义转换；
// 转换器按顺序执行，第一个返回 converted=true 的转换器结果即为最终结果，
// 若所有转换器均未转换，则使用默认逻辑处理。
func OrderByExpr(order clause.OrderBy, convs ...OrderByConverter) gormClause.OrderByColumn {
	var col gormClause.OrderByColumn
	var converted bool
	for _, conv := range convs {
		col, converted = conv(order)
		if converted {
			break
		}
	}
	if !converted {
		col = gormClause.OrderByColumn{
			Column: gormClause.Column{Name: order.Column},
			Desc:   order.Desc,
		}
	}

	return col
}

// OrderByExprs 批量将 clause.OrderBys 转换为 GORM 的 OrderByColumn 列表。
// 转换过程中会跳过 nil 以及列名为空的表达式。
//
// 可传入 OrderByConverter 对特定排序条件进行自定义转换，规则同 OrderByExpr。
func OrderByExprs(orders clause.OrderBys, convs ...OrderByConverter) []gormClause.OrderByColumn {
	cols := make([]gormClause.OrderByColumn, 0, len(orders))
	for _, order := range orders {
		if order == nil {
			continue
		}
		if order.Column == "" {
			continue
		}
		cols = append(cols, OrderByExpr(*order, convs...))
	}

	return cols
}

// OrderByScope 将 clause.OrderBys 转换为 GORM Scope 函数，用于设置排序条件。
//
// 可传入 OrderByConverter 对特定排序条件进行自定义转换。
func OrderByScope(orders clause.OrderBys, convs ...OrderByConverter) func(db *gorm.DB) *gorm.DB {

	return func(db *gorm.DB) *gorm.DB {
		cols := OrderByExprs(orders, convs...)

		if len(cols) > 0 {
			return db.Order(gormClause.OrderBy{Columns: cols})
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
