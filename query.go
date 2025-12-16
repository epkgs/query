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
