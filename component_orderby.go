package query

import (
	"strings"

	"github.com/epkgs/query/clause"
)

// orderbys 是一个通用的排序查询构建器，支持多种排序方式
// P 父 struct，必须实现 errorRecorder 接口
type orderbys[P errorRecorder] struct {
	Parent P
	Value  clause.OrderBys
}

// OrderByExpr 返回当前的排序表达式
func (o *orderbys[P]) OrderByExpr() clause.OrderBys {
	return o.Value
}

// OrderBy 添加排序条件
// 参数:
//   - field: 字段名、表达式或表达式数组
//   - args: 条件参数，格式根据field类型而定
//
// 返回值:
//   - 当前实例，支持链式调用
//
// 示例:
//   - q.OrderBy("name", "desc")             // name DESC
//   - q.OrderBy("age asc")                  // age ASC
//   - q.OrderBy(clause.OrderBy{Column: "name", Desc: true})  // 使用clause.Expression
//   - q.OrderBy([]clause.OrderBy{ {Column: "name", Desc: true}, {Column: "age", Desc: false} }) // 多个排序子句
func (o *orderbys[P]) OrderBy(field any, orders ...any) P {

	switch f := field.(type) {

	case string:
		// 处理字符串类型的排序条件
		if len(orders) > 0 {
			if ord, ok := orders[0].(string); ok {
				// 字段名和方向分别作为参数传入
				orderBys := buildOrderBy(f, ord)
				o.Value = append(o.Value, orderBys...)
				break
			}
		}

		// 字段名包含方向信息（如 "age desc, name asc"）
		orderBys := buildOrderBy(f)
		o.Value = append(o.Value, orderBys...)

	case clause.OrderBy:
		// 处理单个clause.OrderBy
		o.Value = append(o.Value, f)
		// 处理额外的clause.OrderBy参数
		for _, ord := range orders {
			if ord, ok := ord.(clause.OrderBy); ok {
				o.Value = append(o.Value, ord)
			}
		}
	case []clause.OrderBy:
		// 处理[]clause.OrderBy切片
		o.Value = append(o.Value, f...)
	case clause.OrderBys:
		// 处理clause.OrderBys集合
		o.Value = append(o.Value, f...)
	default:
		o.Parent.setError(ErrInvalidOrderBy)
	}

	return o.Parent
}

// Build 构建排序子句
func (o *orderbys[P]) Build(builder clause.Builder) {
	o.Value.Build(builder)
}

// buildOrderBy 解析字符串形式的排序条件，返回clause.OrderBys
// 支持两种形式：
// 1. column string: 单个字段名（默认升序）或包含方向的字段名列表（如 "name asc, age desc"）
// 2. column string, direction string: 字段名和方向（如 "name", "desc"）
func buildOrderBy(column string, direction ...string) clause.OrderBys {

	if len(direction) == 0 {
		// 解析包含方向的字段名列表
		orders := make(clause.OrderBys, 0)

		orderStrs := strings.Split(column, ",")
		for _, orderStr := range orderStrs {
			orderStr = strings.TrimSpace(orderStr)
			if orderStr == "" {
				continue
			}
			pices := strings.Split(orderStr, " ")

			if len(pices) == 0 {
				continue
			}

			// 提取字段名和方向
			fieldName := strings.TrimSpace(pices[0])
			if fieldName == "" {
				continue
			}
			isDesc := false
			if len(pices) > 1 {
				isDesc = strings.ToLower(strings.TrimSpace(pices[1])) == "desc"
			}

			orders = append(orders, clause.OrderBy{
				Column: fieldName,
				Desc:   isDesc,
			})
		}

		return orders
	}

	// 解析字段名和方向作为单独参数的情况
	return clause.OrderBys{
		{
			Column: column,
			Desc:   strings.ToLower(direction[0]) == "desc",
		},
	}
}
