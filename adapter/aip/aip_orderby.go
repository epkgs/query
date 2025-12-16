package aip

import (
	"github.com/epkgs/query/clause"
	ordering "go.einride.tech/aip/ordering"
)

// FromOrderBy 将 aip/ordering.OrderBy 转换为 *clause.OrderBys
// 用于将 AIP 排序指令转换为查询构建器可使用的排序条件
func FromOrderBy(orderBy ordering.OrderBy) clause.OrderBys {
	// 创建空的 clause.OrderBys
	orderBys := clause.OrderBys{}

	// 遍历 ordering.OrderBy 的 Fields
	for _, field := range orderBy.Fields {
		// 将每个 ordering.Field 转换为 clause.OrderBy
		orderBys = append(orderBys, clause.OrderBy{
			Column: field.Path, // Field.Path 映射到 Column
			Desc:   field.Desc, // Field.Desc 映射到 Desc
		})
	}

	return orderBys
}
