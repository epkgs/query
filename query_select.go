package query

import "github.com/epkgs/query/clause"

// SelectQuery SELECT查询结构体
type SelectQuery struct {
	table string
	errorRecord
	fields []string

	*orderbys[*SelectQuery]
	*pagination[*SelectQuery]
	*where[*SelectQuery]
}

// Select 设置SELECT查询的字段
func (q *SelectQuery) Select(fields ...string) *SelectQuery {
	q.fields = append(q.fields, fields...)
	return q
}

// Build 构建SELECT查询的SQL语句
func (q *SelectQuery) Build(builder clause.Builder) {
	// 构建 SELECT 部分
	builder.WriteString("SELECT ")
	if len(q.fields) > 0 {
		for i, field := range q.fields {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteQuoted(field)
		}
	} else {
		builder.WriteString("*")
	}

	// 构建 FROM 部分
	if q.table != "" {
		builder.WriteString(" FROM ")
		builder.WriteQuoted(q.table)
	}

	// 构建 WHERE 部分
	q.where.Build(builder)

	// 构建 ORDER BY 部分
	q.orderbys.Build(builder)

	// 构建 Pagination 部分
	q.pagination.Build(builder)
}
