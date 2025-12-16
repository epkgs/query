package query

import "github.com/epkgs/query/clause"

// UpdateQuery UPDATE查询结构体
type UpdateQuery struct {
	table string
	errorRecord
	*where[*UpdateQuery]
	*pagination[*UpdateQuery]

	values map[string]interface{}
}

// Update 设置UPDATE查询的字段值
// 支持两种调用方式：
// 1. Update("field", value) - 设置单个字段值
// 2. Update(map[string]interface{}) - 设置多个字段值
func (q *UpdateQuery) Update(column interface{}, value ...interface{}) *UpdateQuery {
	switch v := column.(type) {
	case string:
		// Set("field", value) 形式
		if len(value) > 0 {
			q.values[v] = value[0]
		}
	case map[string]interface{}:
		// Set(map[string]interface{}) 形式
		for k, val := range v {
			q.values[k] = val
		}
	}

	return q
}

// Build 构建UPDATE查询的SQL语句
func (q *UpdateQuery) Build(builder clause.Builder) {
	// 构建 UPDATE 部分
	builder.WriteString("UPDATE ")
	builder.WriteQuoted(q.table)

	// 构建 SET 部分
	if len(q.values) > 0 {
		builder.WriteString(" SET ")
		// 对字段进行排序，确保生成的SQL有固定的字段顺序
		var fields []string
		for field := range q.values {
			fields = append(fields, field)
		}
		// 简单排序（按字典序）
		for i := 0; i < len(fields)-1; i++ {
			for j := i + 1; j < len(fields); j++ {
				if fields[i] > fields[j] {
					fields[i], fields[j] = fields[j], fields[i]
				}
			}
		}
		// 构建SET子句
		for i, field := range fields {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteQuoted(field)
			builder.WriteString(" = ")
			builder.AddVar(builder, q.values[field])
		}
	}

	// 构建 WHERE 部分
	q.where.Build(builder)

	// 构建分页部分
	if q.pagination.Value.Limit != nil && *q.pagination.Value.Limit > 0 {
		builder.WriteString(" LIMIT ")
		builder.AddVar(builder, *q.pagination.Value.Limit)
	}
}
