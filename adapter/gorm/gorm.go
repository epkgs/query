package gorm

import (
	"github.com/epkgs/query/clause"
	"gorm.io/gorm"
	gormClause "gorm.io/gorm/clause"
)

// Where 将 clause.Where 转换为 gorm scope
func Where(where clause.Where) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(where.Exprs) == 0 {
			return db
		}

		// 将 query/clause.Where 转换为 gorm/clause.Where
		gormWhere := convertWhere(where)

		// 将 gorm/clause.Where 给到 gorm.DB 的 Where 函数
		return db.Where(gormWhere)
	}
}

// convertWhere 将 query/clause.Where 转换为 gorm/clause.Where
func convertWhere(where clause.Where) gormClause.Where {
	var gormExprs []gormClause.Expression

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

// OrderBy 将 clause.OrderBys 转换为 gorm scope
func OrderBy(orders clause.OrderBys) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(orders) == 0 {
			return db
		}

		gOrderByCols := []gormClause.OrderByColumn{}

		for _, order := range orders {
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
func Query(where clause.Where, orders clause.OrderBys, pagination clause.Pagination) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db = Where(where)(db)
		db = OrderBy(orders)(db)
		db = Pagination(pagination)(db)
		return db
	}
}
