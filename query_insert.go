package query

import "github.com/epkgs/query/clause"

// InsertQuery SELECT查询结构体
type InsertQuery struct {
	table string
	errorRecord

	values []map[string]any
}

// Insert 设置INSERT查询的插入值
// 支持三种调用方式：
// 1. Insert("field", value) - 设置单行单个字段值
// 2. Insert(map[string]any) - 插入单行数据
// 3. Insert(map1, map2) - 插入多行数据
func (q *InsertQuery) Insert(field any, args ...any) *InsertQuery {

	// 处理 Insert(map[string]any) 或 Insert(map1, map2) 形式
	if m, ok := field.(map[string]any); ok {
		q.values = append(q.values, m)
		// 检查是否是多行插入
		if len(args) > 0 {
			// Insert(map1, map2) 形式
			for _, arg := range args {
				if row, ok := arg.(map[string]any); ok {
					q.values = append(q.values, row)
				}
			}
		}
		return q
	}

	allArgs := append([]any{field}, args...)

	// 处理 Insert("field", value) 形式
	if len(allArgs)%2 == 0 {
		row := make(map[string]any)
		for i := 0; i < len(allArgs); i += 2 {
			if field, ok := allArgs[i].(string); ok {
				row[field] = allArgs[i+1]
			}
		}
		if len(row) > 0 {
			q.values = append(q.values, row)
		}
	} else {
		q.Error = ErrInvalidInsertValues
	}

	return q
}

// Build 构建INSERT查询的SQL语句
func (q *InsertQuery) Build(builder clause.Builder) {
	// 构建 INSERT 部分
	builder.WriteString("INSERT INTO ")
	builder.WriteQuoted(q.table)

	// 构建 VALUES 部分
	if len(q.values) > 0 {
		// 获取所有字段名
		var allFields []string
		fieldMap := make(map[string]bool)
		for _, row := range q.values {
			for field := range row {
				if !fieldMap[field] {
					fieldMap[field] = true
					allFields = append(allFields, field)
				}
			}
		}

		// 对字段进行排序，确保生成的SQL有固定的字段顺序
		for i := 0; i < len(allFields)-1; i++ {
			for j := i + 1; j < len(allFields); j++ {
				if allFields[i] > allFields[j] {
					allFields[i], allFields[j] = allFields[j], allFields[i]
				}
			}
		}

		// 写入字段名
		builder.WriteString(" (")
		for i, field := range allFields {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteQuoted(field)
		}
		builder.WriteString(") VALUES ")

		// 写入值
		for i, row := range q.values {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString("(")
			for j, field := range allFields {
				if j > 0 {
					builder.WriteString(", ")
				}
				// 检查字段是否存在于当前行，如果不存在则使用nil
				if value, ok := row[field]; ok {
					builder.AddVar(builder, value)
				} else {
					builder.AddVar(builder, nil)
				}
			}
			builder.WriteString(")")
		}
	}
}
