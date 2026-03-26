package query

import (
	"strings"

	"github.com/epkgs/query/clause"
)

type genericSorter[Q any] interface {
	Asc(column string) Q
	Desc(column string) Q
	OrderBy(field any, orders ...any) Q
	OrderByExpr(walker ...OrderByWalker) clause.OrderBys
}

var _ genericSorter[*Query] = (*orderbys[*Query])(nil)
var _ clause.Expression = (*orderbys[*Query])(nil)

type orderbyExprExporter interface {
	OrderByExpr(walker ...OrderByWalker) clause.OrderBys
}

type orderbyQuerier interface {
	errorRecorder
	orderbyExprExporter
}

// orderbys 是一个通用的排序查询构建器，支持多种排序方式
// Q 是一个实现了 orderbyQuerier 接口的查询类型，通常是 *Query
type orderbys[Q orderbyQuerier] struct {
	Parent Q
	Value  clause.OrderBys
}

// OrderByExpr 返回当前的排序表达式
func (o *orderbys[Q]) OrderByExpr(walker ...OrderByWalker) clause.OrderBys {
	return walkOrderByExpr(o.Value, walker...)
}

// OrderByWalker 定义一个函数类型，用于遍历和修改 *clause.OrderBy， 返回 nil 则表示删除该表达式
type OrderByWalker func(*clause.OrderBy) *clause.OrderBy

func walkOrderByExpr(orderbys clause.OrderBys, walker ...OrderByWalker) clause.OrderBys {
	if len(walker) == 0 {
		return orderbys
	}

	copied := make(clause.OrderBys, len(orderbys))

	copy(copied, orderbys)

	result := make(clause.OrderBys, 0, len(orderbys))

	for _, w := range walker {
		for _, ob := range copied {
			if newOb := w(ob); newOb != nil {
				result = append(result, newOb)
			}
		}
	}

	return result
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
func (o *orderbys[Q]) OrderBy(field any, orders ...any) Q {

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
		o.Value = append(o.Value, &f)
		// 处理额外的clause.OrderBy参数
		if len(orders) > 0 {
			o.OrderBy(orders[0], orders[1:]...)
		}

	case []clause.OrderBy:
		// 处理[]clause.OrderBy切片
		for _, ob := range f {
			newOb := ob // 先赋值、在传引用。否则直接传引用会导致指针指向最后一个元素的地址
			o.Value = append(o.Value, &newOb)
		}
	case *clause.OrderBy:
		// 处理单个clause.OrderBy
		o.Value = append(o.Value, f)
		// 处理额外的clause.OrderBy参数
		if len(orders) > 0 {
			o.OrderBy(orders[0], orders[1:]...)
		}

	case []*clause.OrderBy:
		// 处理[]*clause.OrderBy切片
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
func (o *orderbys[Q]) Build(builder clause.Builder) {
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

			orders = append(orders, &clause.OrderBy{
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

// ========== 流畅链式 API (Fluent Chain API) ==========
// 类似 GORM 的链式调用方式，无需显式调用 Build()

// Desc 添加降序排序
func (o *orderbys[Q]) Desc(column string) Q {
	o.Value = append(o.Value, &clause.OrderBy{Column: column, Desc: true})
	return o.Parent
}

// Asc 添加升序排序
func (o *orderbys[Q]) Asc(column string) Q {
	o.Value = append(o.Value, &clause.OrderBy{Column: column, Desc: false})
	return o.Parent
}

// Desc 添加降序排序
func Desc(column string) *Query {
	return newQuery("").Desc(column)
}

// Asc 添加升序排序
func Asc(column string) *Query {
	return newQuery("").Asc(column)
}
