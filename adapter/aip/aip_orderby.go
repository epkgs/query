package aip

import (
	"github.com/epkgs/query/clause"
	ordering "go.einride.tech/aip/ordering"
)

// FromOrderBy 将 AIP 标准的 ordering.OrderBy 转换为 clause.OrderBys。
// ordering.OrderBy（来自 go.einride.tech/aip/ordering 包）的每个 Field
// 会被映射为 clause.OrderBy，其中 Field.Path 对应 Column，Field.Desc 对应 Desc。
//
// 示例：
//
//	// 假设客户端请求 order_by="name desc, age asc"
//	parsed, _ := ordering.ParseOrderBy("name desc, age asc")
//	orderBys := FromOrderBy(parsed)
//	// orderBys 包含两个排序条件：name DESC 和 age ASC
func FromOrderBy(orderBy ordering.OrderBy) clause.OrderBys {
	// 创建空的 clause.OrderBys
	orderBys := clause.OrderBys{}

	// 遍历 ordering.OrderBy 的 Fields
	for _, field := range orderBy.Fields {
		// 将每个 ordering.Field 转换为 clause.OrderBy
		orderBys = append(orderBys, &clause.OrderBy{
			Column: field.Path, // Field.Path 映射到 Column
			Desc:   field.Desc, // Field.Desc 映射到 Desc
		})
	}

	return orderBys
}
