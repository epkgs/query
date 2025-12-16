package query

import "github.com/epkgs/query/clause"

// DeleteQuery DELETE查询结构体
type DeleteQuery struct {
	table string
	errorRecord
	*where[*DeleteQuery]
}

// Build 构建DELETE查询的SQL语句
func (q *DeleteQuery) Build(builder clause.Builder) {
	// 构建 DELETE 部分
	builder.WriteString("DELETE FROM ")
	builder.WriteQuoted(q.table)

	// 构建 WHERE 部分
	q.where.Build(builder)
}
