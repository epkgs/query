// Package query 是一个轻量级、无依赖的 Go 语言 SQL 查询构建器。
//
// 它提供流畅的链式调用 API，支持 SELECT、INSERT、UPDATE、DELETE 操作，
// 以及复杂的 WHERE 条件（AND/OR/NOT）、ORDER BY 排序和分页。
//
// 该包本身只构建抽象查询表达式，可通过适配器（adapter 子包）将查询
// 转换为 GORM、Ent 等 ORM 的查询条件，或转换为标准 SQL 语句。
//
// 快速开始:
//
//	// SELECT 查询
//	q := query.Table("users").
//	    Where("age", ">", 18).
//	    OrderBy("name", "desc").
//	    Limit(10).Select("id", "name", "age")
//
//	// INSERT 操作
//	q := query.Table("users").Insert(map[string]any{"name": "John", "age": 30})
//
//	// UPDATE 操作
//	q := query.Table("users").Where("id", 1).Update("name", "John")
//
//	// DELETE 操作
//	q := query.Table("users").Where("id", 1).Delete()
//
// AIP 与 ORM 集成:
//
//	import (
//	    query "github.com/epkgs/query"
//	    gormadapter "github.com/epkgs/query/adapter/gorm"
//	    aipfilter "go.einride.tech/aip/filtering"
//	)
//
//	filter, _ := aipfilter.ParseFilter(request, declarations)
//	whereClause, _ := aipadapter.FromFilter(filter)
//	db.Scopes(gormadapter.WhereScope(whereClause)).Find(&users)
package query

import (
	"errors"

	"github.com/epkgs/query/clause"
)

var (
	ErrInvalidOperator     = errors.New("invalid operator")
	ErrInvalidCondition    = errors.New("invalid condition")
	ErrInvalidInsertValues = errors.New("invalid insert values")
	ErrInvalidOrderBy      = errors.New("invalid order by")
)

var _ genericWherer[*Query] = (*Query)(nil)

// Query 是查询构建器的基础结构体。
// 通过 Table() 函数创建实例，然后调用链式方法构建 WHERE 条件、
// ORDER BY 排序和分页参数，最后调用 Select/Insert/Update/Delete 转换为具体操作。
type Query struct {
	table string

	errorRecord
	*where[*Query]
	*orderbys[*Query]
	*pagination[*Query]
}

func newQuery(tableName string) *Query {
	q := &Query{
		table: tableName,
	}

	q.where = &where[*Query]{
		Parent: q,
		Value:  clause.Where{},
	}

	q.pagination = &pagination[*Query]{
		Parent: q,
		Value:  clause.Pagination{},
	}

	q.orderbys = &orderbys[*Query]{
		Parent: q,
		Value:  clause.OrderBys{},
	}

	return q
}

// Table 创建一个新的查询构建器并设置表名。
// 这是所有查询操作的入口函数。
//
// 示例:
//
//	// 基本 SELECT
//	q := query.Table("users").Eq("status", "active").Select("id", "name")
//
//	// INSERT
//	q := query.Table("users").Insert("name", "John")
//
//	// UPDATE
//	q := query.Table("users").Where("id", 1).Update("name", "John")
//
//	// DELETE
//	q := query.Table("users").Where("id", 1).Delete()
func Table(tableName string) *Query {
	return newQuery(tableName)
}

// Where 添加WHERE条件到当前查询
//
// 建议使用 Eq, Neq, Gt 等方法替代，例如：q.Eq("name", "John")
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
//   - q.Where(func(w Wherer) Wherer {  w.Where("name", "John"); return w })
func Where(field any, args ...any) *Query {
	return newQuery("").Where(field, args...)
}

// OrWhere 添加OR WHERE条件到当前查询
//
// 建议使用 Or 方法替代，例如：q.Or("name", "John")
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
//   - q.OrWhere(func(w Wherer) Wherer {  w.Where("name", "John"); return w })
func OrWhere(field any, args ...any) *Query {
	return newQuery("").OrWhere(field, args...)
}

// OrderBy 添加ORDER BY子句到当前查询
// 参数:
//   - field: 字段名、表达式或表达式数组
//   - args: 排序参数，格式根据field类型而定
//
// 返回值:
//   - 当前实例，支持链式调用
//
// 示例:
//   - q.OrderBy("age")
//   - q.OrderBy("age", "desc")
//   - q.OrderBy(clause.OrderBy{Column: "age", Direction: clause.Desc})
//   - q.OrderBy(func(o OrderByser) OrderByser {  o.OrderBy("age"); return o })
func OrderBy(field any, args ...any) *Query {
	return newQuery("").OrderBy(field, args...)
}

// Limit 设置查询的限制条数
func Limit(limit int) *Query {
	return newQuery("").Limit(limit)
}

// Offset 设置查询的偏移量
func Offset(offset int) *Query {
	return newQuery("").Offset(offset)
}

// Paginate 设置分页参数
func Paginate(page, pageSize int) *Query {
	return newQuery("").Paginate(page, pageSize)
}

// Table 设置查询的表名
func (q *Query) Table(tableName string) *Query {
	q.table = tableName
	return q
}

// Select 将查询转换为 SELECT 查询并指定要查询的字段。
// 此方法将 *Query 转换为 *SelectQuery，继承当前查询的 WHERE 条件、
// ORDER BY 排序和分页参数。
//
// 如果不指定字段，将默认选择所有字段 (*)。
//
// 示例:
//
//	q.Select("id", "name", "email")  // SELECT id, name, email FROM ...
//	q.Select()                        // SELECT * FROM ...
func (q *Query) Select(fields ...string) *SelectQuery {
	sq := &SelectQuery{
		table:       q.table,
		errorRecord: q.errorRecord,

		fields: fields,
	}

	sq.where = &where[*SelectQuery]{
		Parent: sq,
		Value:  q.where.Value,
	}

	sq.pagination = &pagination[*SelectQuery]{
		Parent: sq,
		Value:  q.pagination.Value,
	}

	sq.orderbys = &orderbys[*SelectQuery]{
		Parent: sq,
		Value:  q.orderbys.Value,
	}
	return sq
}

// Insert 设置INSERT查询的插入值
// 支持三种调用方式：
// 1. Insert("field", value) - 设置单行单个字段值
// 2. Insert(map[string]any) - 插入单行数据
// 3. Insert(map1, map2) - 插入多行数据
func (q *Query) Insert(field any, args ...any) *InsertQuery {
	query := &InsertQuery{
		table:       q.table,
		errorRecord: q.errorRecord,
		values:      make([]map[string]any, 0),
	}
	return query.Insert(field, args...)
}

// Update 设置UPDATE查询的字段值
// 支持两种调用方式：
// 1. Update("field", value) - 设置单个字段值
// 2. Update(map[string]any) - 设置多个字段值
func (q *Query) Update(column any, value ...any) *UpdateQuery {
	query := &UpdateQuery{
		table:       q.table,
		errorRecord: q.errorRecord,
		values:      make(map[string]any),
	}

	query.where = &where[*UpdateQuery]{
		Parent: query,
		Value:  q.where.Value,
	}

	query.pagination = &pagination[*UpdateQuery]{
		Parent: query,
		Value:  q.pagination.Value,
	}

	return query.Update(column, value...)
}

// Delete 将查询转换为DELETE查询
func (q *Query) Delete() *DeleteQuery {
	query := &DeleteQuery{
		table:       q.table,
		errorRecord: q.errorRecord,
	}

	query.where = &where[*DeleteQuery]{
		Parent: query,
		Value:  q.where.Value,
	}

	return query
}
