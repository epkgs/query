package gorm

import (
	"testing"

	"github.com/epkgs/query"
	"github.com/epkgs/query/clause"
	"gorm.io/gorm"
)

// 测试基本的 Where 条件转换
func TestWhereScope(t *testing.T) {
	// 创建查询条件
	q := query.Table("").Where("name", "John").Where("age", ">", 18)

	// 获取 Where 表达式
	where := q.WhereExpr()

	// 转换为 gorm scope
	scope := Where(where)

	// 创建 gorm DB 实例
	db, _ := gorm.Open(nil, &gorm.Config{})

	// 应用 scope
	db = scope(db)

	// 检查生成的 SQL
	t.Logf("SQL: %s", db.Statement.SQL.String())
}

// 测试 OrderBy 条件转换
func TestOrderByScope(t *testing.T) {
	// 创建 OrderBy 条件
	var orders clause.OrderBys
	orders = append(orders, clause.OrderBy{Column: "name", Desc: false})
	orders = append(orders, clause.OrderBy{Column: "age", Desc: true})

	// 转换为 gorm scope
	scope := OrderBy(orders)

	// 创建 gorm DB 实例
	db, _ := gorm.Open(nil, &gorm.Config{})

	// 应用 scope
	db = scope(db)

	// 检查生成的 SQL
	t.Logf("SQL: %s", db.Statement.SQL.String())
}

// 测试 Pagination 条件转换
func TestPaginationScope(t *testing.T) {
	// 创建 Pagination 条件
	limit := 10
	pagination := clause.Pagination{
		Limit:  &limit,
		Offset: 20,
	}

	// 转换为 gorm scope
	scope := Pagination(pagination)

	// 创建 gorm DB 实例
	db, _ := gorm.Open(nil, &gorm.Config{})

	// 应用 scope
	db = scope(db)

	// 检查生成的 SQL
	t.Logf("SQL: %s", db.Statement.SQL.String())
}

// 测试完整的 Query 条件转换
func TestQueryScope(t *testing.T) {
	// 创建 Where 条件
	q := query.Table("").Where("name", "John").Where("age", ">", 18)
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

	// 转换为 gorm scope
	scope := Query(where, orders, pagination)

	// 创建 gorm DB 实例
	db, _ := gorm.Open(nil, &gorm.Config{})

	// 应用 scope
	db = scope(db)

	// 检查生成的 SQL
	t.Logf("SQL: %s", db.Statement.SQL.String())
}

// 测试 Or 条件转换
func TestOrWhereScope(t *testing.T) {
	// 创建查询条件
	q := query.Table("").OrWhere(func(w query.Wherer) query.Wherer {
		w.Where("name", "John")
		w.Where("age", 30)
		return w
	})
	q.OrWhere("city", "New York")

	// 获取 Where 表达式
	where := q.WhereExpr()

	// 转换为 gorm scope
	scope := Where(where)

	// 创建 gorm DB 实例
	db, _ := gorm.Open(nil, &gorm.Config{})

	// 应用 scope
	db = scope(db)

	// 检查生成的 SQL
	t.Logf("SQL: %s", db.Statement.SQL.String())
}

// 测试 Not 条件转换
func TestNotWhereScope(t *testing.T) {
	// 创建查询条件
	q := query.Table("").Not("name", "John").Not("age", ">", 18)

	// 获取 Where 表达式
	where := q.WhereExpr()

	// 转换为 gorm scope
	scope := Where(where)

	// 创建 gorm DB 实例
	db, _ := gorm.Open(nil, &gorm.Config{})

	// 应用 scope
	db = scope(db)

	// 检查生成的 SQL
	t.Logf("SQL: %s", db.Statement.SQL.String())
}

// 测试 IN 条件转换
func TestInWhereScope(t *testing.T) {
	// 创建查询条件
	q := query.Table("").Where("id", "IN", []interface{}{1, 2, 3})

	// 获取 Where 表达式
	where := q.WhereExpr()

	// 转换为 gorm scope
	scope := Where(where)

	// 创建 gorm DB 实例
	db, _ := gorm.Open(nil, &gorm.Config{})

	// 应用 scope
	db = scope(db)

	// 检查生成的 SQL
	t.Logf("SQL: %s", db.Statement.SQL.String())
}
