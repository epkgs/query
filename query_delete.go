package query

import "github.com/epkgs/query/clause"

// DeleteQuery 是 DELETE 查询结构体。
// 包含表名和 WHERE 条件，通过 Build 方法将完整的 DELETE 语句写入 Builder。
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
