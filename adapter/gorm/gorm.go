package gorm

import (
	"github.com/epkgs/query/clause"
	"gorm.io/gorm"
	gormClause "gorm.io/gorm/clause"
)

type options struct {
	mapper MapperFunc
}

type Option func(*options)

// MapperFunc 定义了查询列名到 GORM 列名的映射函数类型
// 输入参数 queryColumn 是查询条件中使用的列名
// 返回值 gormColumn 是数据库表中的实际列名
type MapperFunc func(queryColumn string) (gormColumn string)

// WithMapper 创建一个配置选项，用于设置查询列名到数据库实际列名的映射函数
// mapper 参数是一个 MapperFunc 类型的函数，用于自定义列名映射逻辑
// 该选项用于在将查询条件转换为 GORM 条件时，通过自定义函数实现灵活的列名映射
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
	var gormExprs []gormClause.Expression

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
	switch e := expr.(type) {
	case clause.Eq:
		return gormClause.Eq{Column: gormClause.Column{Name: opt.Column(e.Column)}, Value: e.Value}
	case clause.Neq:
		return gormClause.Neq{Column: gormClause.Column{Name: opt.Column(e.Column)}, Value: e.Value}
	case clause.Gt:
		return gormClause.Gt{Column: gormClause.Column{Name: opt.Column(e.Column)}, Value: e.Value}
	case clause.Gte:
		return gormClause.Gte{Column: gormClause.Column{Name: opt.Column(e.Column)}, Value: e.Value}
	case clause.Lt:
		return gormClause.Lt{Column: gormClause.Column{Name: opt.Column(e.Column)}, Value: e.Value}
	case clause.Lte:
		return gormClause.Lte{Column: gormClause.Column{Name: opt.Column(e.Column)}, Value: e.Value}
	case clause.Like:
		return gormClause.Like{Column: gormClause.Column{Name: opt.Column(e.Column)}, Value: e.Value}
	case clause.IN:
		return gormClause.IN{Column: gormClause.Column{Name: opt.Column(e.Column)}, Values: e.Values}
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
			gOrderByCols = append(gOrderByCols, gormClause.OrderByColumn{
				Column: gormClause.Column{Name: opt.Column(order.Column)},
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
