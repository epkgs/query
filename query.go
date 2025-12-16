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

// Query 基础查询结构体
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

// Table 设置查询的表名
func Table(tableName string) *Query {
	return newQuery(tableName)
}

// Where 添加WHERE条件到当前查询
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

// Not 添加NOT条件到当前查询
// 参数:
//   - field: 字段名、表达式或表达式数组
//   - args: 条件参数，格式根据field类型而定
//
// 返回值:
//   - 当前实例，支持链式调用
//
// 示例:
//   - q.Not("name", "John")
//   - q.Not("age", ">", 18)
//   - q.Not(clause.Eq{Column: "name", Value: "John"})
//   - q.Not(func(w Wherer) Wherer {  w.Where("name", "John"); return w })
func Not(field any, args ...any) *Query {
	return newQuery("").Not(field, args...)
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

// Select 将查询转换为SELECT查询并设置字段
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
