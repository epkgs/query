package ent

import (
	"testing"

	"entgo.io/ent/dialect/sql"
	"github.com/epkgs/query"
	"github.com/epkgs/query/clause"
)

// 测试基本的 Where 条件转换
func TestWhereFunc(t *testing.T) {
	// 创建查询条件
	q := query.Table("").Where("name", "John").Where("age", ">", 18)

	// 获取 Where 表达式
	where := q.WhereExpr()

	// 转换为 ent Where 函数
	whereFunc := Where(where)

	// 创建 ent Selector
	selector := sql.Select("*").From(sql.Table("users"))

	// 应用 Where 函数
	whereFunc(selector)

	// 检查生成的 SQL
	sqlStr, args := selector.Query()
	t.Logf("SQL: %s, Args: %v", sqlStr, args)
}

// 测试 OrderBy 条件转换
func TestOrderBy(t *testing.T) {
	// 创建 OrderBy 条件
	var orders clause.OrderBys
	orders = append(orders, clause.OrderBy{Column: "name", Desc: false})
	orders = append(orders, clause.OrderBy{Column: "age", Desc: true})

	// 转换为 ent OrderBy 函数
	orderByFunc := OrderBy(orders)

	// 创建 ent Selector
	selector := sql.Select("*").From(sql.Table("users"))

	// 应用 OrderBy 函数
	orderByFunc(selector)

	// 检查生成的 SQL
	sqlStr, args := selector.Query()
	t.Logf("SQL: %s, Args: %v", sqlStr, args)
}

// 测试 Pagination 条件转换
func TestPagination(t *testing.T) {
	// 创建 Pagination 条件
	limit := 10
	pagination := clause.Pagination{
		Limit:  &limit,
		Offset: 20,
	}

	// 转换为 ent Pagination 函数
	paginationFunc := Pagination(pagination)

	// 创建 ent Selector
	selector := sql.Select("*").From(sql.Table("users"))

	// 应用 Pagination 函数
	paginationFunc(selector)

	// 检查生成的 SQL
	sqlStr, args := selector.Query()
	t.Logf("SQL: %s, Args: %v", sqlStr, args)
}

// 测试完整的 Query 条件转换
func TestQuery(t *testing.T) {
	// 创建 Where 条件
	q := query.Table("").Where("name", "John").
		Where("age", ">", 18)
	where := q.WhereExpr()

	// 创建 OrderBy 条件
	var orders clause.OrderBys
	orders = append(orders, clause.OrderBy{Column: "name", Desc: false})
	orders = append(orders, clause.OrderBy{Column: "age", Desc: true})

	// 创建 Pagination 条件
	limit := 10
	pagination := clause.Pagination{
		Limit:  &limit,
		Offset: 20,
	}

	// 转换为 ent Query 函数
	queryFunc := Query(where, orders, pagination)

	// 创建 ent Selector
	selector := sql.Select("*").From(sql.Table("users"))

	// 应用 Query 函数
	queryFunc(selector)

	// 检查生成的 SQL
	sqlStr, args := selector.Query()
	t.Logf("SQL: %s, Args: %v", sqlStr, args)
}

// 测试 Or 条件转换
func TestOrWhereFunc(t *testing.T) {
	// 创建查询条件
	q := query.Table("").OrWhere(func(w query.Wherer) query.Wherer {
		w.Where("name", "John")
		w.Where("age", 30)
		return w
	})
	q.OrWhere("city", "New York")

	// 获取 Where 表达式
	where := q.WhereExpr()

	// 转换为 ent Where 函数
	whereFunc := Where(where)

	// 创建 ent Selector
	selector := sql.Select("*").From(sql.Table("users"))

	// 应用 Where 函数
	whereFunc(selector)

	// 检查生成的 SQL
	sqlStr, args := selector.Query()
	t.Logf("SQL: %s, Args: %v", sqlStr, args)
}

// 测试 Not 条件转换
func TestNotWhereFunc(t *testing.T) {
	// 创建查询条件
	q := query.Table("").Not("name", "John").Not("age", ">", 18)

	// 获取 Where 表达式
	where := q.WhereExpr()

	// 转换为 ent Where 函数
	whereFunc := Where(where)

	// 创建 ent Selector
	selector := sql.Select("*").From(sql.Table("users"))

	// 应用 Where 函数
	whereFunc(selector)

	// 检查生成的 SQL
	sqlStr, args := selector.Query()
	t.Logf("SQL: %s, Args: %v", sqlStr, args)
}

// 测试 IN 条件转换
func TestInWhereFunc(t *testing.T) {
	// 创建查询条件
	wherer := query.Table("").Where("id", "IN", []any{1, 2, 3})

	// 获取 Where 表达式
	where := wherer.WhereExpr()

	// 转换为 ent Where 函数
	whereFunc := Where(where)

	// 创建 ent Selector
	selector := sql.Select("*").From(sql.Table("users"))

	// 应用 Where 函数
	whereFunc(selector)

	// 检查生成的 SQL
	sqlStr, args := selector.Query()
	t.Logf("SQL: %s, Args: %v", sqlStr, args)
}
