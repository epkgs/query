package ent

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"entgo.io/ent/dialect/sql"
	"github.com/epkgs/query"
	"github.com/epkgs/query/clause"
)

// 测试基本的 Where 条件转换
func TestWhereFunc(t *testing.T) {
	// 创建查询条件
	q := query.Where("name", "John").Where("age", ">", 18)

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
	expectedSQL := "SELECT * FROM `users` WHERE `name` = ? AND `age` > ?"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// 验证参数
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
		return
	}
	if args[0] != "John" {
		t.Errorf("Expected first arg to be 'John', got %v", args[0])
	}
	if args[1] != 18 {
		t.Errorf("Expected second arg to be 18, got %v", args[1])
	}
}

// 测试空 Where 条件
func TestWhereFunc_Empty(t *testing.T) {
	// 创建空的查询条件
	where := clause.Where{}

	// 转换为 ent Where 函数
	whereFunc := Where(where)

	// 创建 ent Selector
	selector := sql.Select("*").From(sql.Table("users"))

	// 应用 Where 函数
	whereFunc(selector)

	// 检查生成的 SQL
	sqlStr, _ := selector.Query()
	expectedSQL := "SELECT * FROM `users`"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// 空条件不应该有 WHERE 子句
	if strings.Contains(sqlStr, "WHERE") {
		t.Error("Expected no WHERE clause for empty conditions")
	}
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
	sqlStr, _ := selector.Query()
	expectedSQL := "SELECT * FROM `users` ORDER BY `name` ASC, `age` DESC"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// 验证 ORDER BY 子句
	if !strings.Contains(sqlStr, "ORDER BY") {
		t.Error("Expected ORDER BY clause in SQL")
	}
	if !strings.Contains(sqlStr, "name") {
		t.Error("Expected name in ORDER BY clause")
	}
	if !strings.Contains(sqlStr, "age") {
		t.Error("Expected age in ORDER BY clause")
	}
	if !strings.Contains(sqlStr, "DESC") {
		t.Error("Expected DESC in ORDER BY clause")
	}
}

// 测试空 OrderBy 条件
func TestOrderBy_Empty(t *testing.T) {
	// 创建空的 OrderBy 条件
	var orders clause.OrderBys

	// 转换为 ent OrderBy 函数
	orderByFunc := OrderBy(orders)

	// 创建 ent Selector
	selector := sql.Select("*").From(sql.Table("users"))

	// 应用 OrderBy 函数
	orderByFunc(selector)

	// 检查生成的 SQL
	sqlStr, _ := selector.Query()
	expectedSQL := "SELECT * FROM `users`"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// 空条件不应该有 ORDER BY 子句
	if strings.Contains(sqlStr, "ORDER BY") {
		t.Error("Expected no ORDER BY clause for empty conditions")
	}
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
	sqlStr, _ := selector.Query()
	expectedSQL := "SELECT * FROM `users` LIMIT 10 OFFSET 20"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// 验证 LIMIT 和 OFFSET
	if !strings.Contains(sqlStr, "LIMIT") {
		t.Error("Expected LIMIT in SQL")
	}
	if !strings.Contains(sqlStr, "OFFSET") {
		t.Error("Expected OFFSET in SQL")
	}
	if !strings.Contains(sqlStr, "10") {
		t.Error("Expected LIMIT value 10 in SQL")
	}
	if !strings.Contains(sqlStr, "20") {
		t.Error("Expected OFFSET value 20 in SQL")
	}
}

// 测试只有 Limit 的 Pagination
func TestPagination_LimitOnly(t *testing.T) {
	// 创建 Pagination 条件
	limit := 5
	pagination := clause.Pagination{
		Limit:  &limit,
		Offset: 0,
	}

	// 转换为 ent Pagination 函数
	paginationFunc := Pagination(pagination)

	// 创建 ent Selector
	selector := sql.Select("*").From(sql.Table("users"))

	// 应用 Pagination 函数
	paginationFunc(selector)

	// 检查生成的 SQL
	sqlStr, _ := selector.Query()
	expectedSQL := "SELECT * FROM `users` LIMIT 5"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// 验证 LIMIT
	if !strings.Contains(sqlStr, "LIMIT") {
		t.Error("Expected LIMIT in SQL")
	}
	if !strings.Contains(sqlStr, "5") {
		t.Error("Expected LIMIT value 5 in SQL")
	}
	// 没有 OFFSET 0，通常不会出现在 SQL 中
}

// 测试只有 Offset 的 Pagination
func TestPagination_OffsetOnly(t *testing.T) {
	// 创建 Pagination 条件
	pagination := clause.Pagination{
		Limit:  nil,
		Offset: 10,
	}

	// 转换为 ent Pagination 函数
	paginationFunc := Pagination(pagination)

	// 创建 ent Selector
	selector := sql.Select("*").From(sql.Table("users"))

	// 应用 Pagination 函数
	paginationFunc(selector)

	// 检查生成的 SQL
	sqlStr, _ := selector.Query()
	expectedSQL := "SELECT * FROM `users` OFFSET 10"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// 验证 OFFSET
	if !strings.Contains(sqlStr, "OFFSET") {
		t.Error("Expected OFFSET in SQL")
	}
	if !strings.Contains(sqlStr, "10") {
		t.Error("Expected OFFSET value 10 in SQL")
	}
}

// 测试完整的 Query 条件转换
func TestQuery(t *testing.T) {
	// 创建 Where 条件
	q := query.Where("name", "John").
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
	expectedSQL := "SELECT * FROM `users` WHERE `name` = ? AND `age` > ? ORDER BY `name` ASC, `age` DESC LIMIT 10 OFFSET 20"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// 验证所有组件都存在
	if !strings.Contains(sqlStr, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}
	if !strings.Contains(sqlStr, "ORDER BY") {
		t.Error("Expected ORDER BY clause in SQL")
	}
	if !strings.Contains(sqlStr, "LIMIT") {
		t.Error("Expected LIMIT in SQL")
	}
	if !strings.Contains(sqlStr, "OFFSET") {
		t.Error("Expected OFFSET in SQL")
	}

	// 验证参数
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
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
	expectedSQL := "SELECT * FROM `users` WHERE (`name` = ? AND `age` = ?) OR `city` = ?"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// 验证条件存在（注意：ent 的 SQL 生成可能与我们的 query builder 略有不同）
	if !strings.Contains(sqlStr, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}

	// 验证参数
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}
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
	expectedSQL := "SELECT * FROM `users` WHERE (NOT (`name` = ?)) AND (NOT (`age` > ?))"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// 验证 NOT 条件
	if !strings.Contains(sqlStr, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}
	if !strings.Contains(sqlStr, "NOT") {
		t.Error("Expected NOT in SQL")
	}

	// 验证参数
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
}

// 测试 IN 条件转换
func TestInWhereFunc(t *testing.T) {
	// 创建查询条件
	wherer := query.Where("id", "IN", []any{1, 2, 3})

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
	expectedSQL := "SELECT * FROM `users` WHERE `id` IN (?, ?, ?)"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// 验证 IN 条件
	if !strings.Contains(sqlStr, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}
	if !strings.Contains(sqlStr, "IN") {
		t.Error("Expected IN in SQL")
	}

	// 验证参数
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}
	expectedArgs := []any{1, 2, 3}
	if !reflect.DeepEqual(args, expectedArgs) {
		t.Errorf("Expected args %v, got %v", expectedArgs, args)
	}
}

// 测试 LIKE 条件转换
func TestLikeWhereFunc(t *testing.T) {
	// 创建查询条件
	q := query.Where("name", "LIKE", "%John%")

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
	expectedSQL := "SELECT * FROM `users` WHERE `name` LIKE ?"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// 验证 LIKE 条件
	if !strings.Contains(sqlStr, "LIKE") {
		t.Error("Expected LIKE in SQL")
	}

	// 验证参数
	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
		return
	}
	if args[0] != "%John%" {
		t.Errorf("Expected arg to be '%%John%%', got %v", args[0])
	}
}

// 测试所有比较操作符
func TestComparisonOperators(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		value    any
		sqlPart  string
	}{
		{"Equal", "=", "John", "="},
		{"NotEqual", "!=", "John", "<>"},
		{"GreaterThan", ">", 18, ">"},
		{"GreaterOrEqual", ">=", 18, ">="},
		{"LessThan", "<", 65, "<"},
		{"LessOrEqual", "<=", 65, "<="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 根据操作符选择字段
			field := "name"
			if tt.operator != "=" && tt.operator != "!=" {
				field = "age"
			}

			// 创建查询条件
			q := query.Where(field, tt.operator, tt.value)

			// 获取 Where 表达式
			where := q.WhereExpr()

			// 转换为 ent Where 函数
			whereFunc := Where(where)

			// 创建 ent Selector
			selector := sql.Select("*").From(sql.Table("users"))

			// 应用 Where 函数
			whereFunc(selector)

			// 检查生成的 SQL
			sqlStr, _ := selector.Query()

			t.Logf("SQL: %s", sqlStr)

			if !strings.Contains(sqlStr, tt.sqlPart) {
				t.Errorf("Expected SQL to contain '%s', got: %s", tt.sqlPart, sqlStr)
			}
		})
	}
}

// 测试 ExprHandler 选项
func TestWithExprHandler(t *testing.T) {
	// 创建查询条件
	q := query.Where("name", "John").Where("age", ">", 18)
	where := q.WhereExpr()

	// 创建一个表达式处理器，过滤掉 age 条件
	handler := func(expr clause.Expression) clause.Expression {
		switch e := expr.(type) {
		case clause.Gt:
			if e.Column == "age" {
				return nil // 过滤掉 age 条件
			}
		}
		return expr
	}

	// 转换为 ent Where 函数
	whereFunc := Where(where, WithExprHandler(handler))

	// 创建 ent Selector
	selector := sql.Select("*").From(sql.Table("users"))

	// 应用 Where 函数
	whereFunc(selector)

	// 检查生成的 SQL
	sqlStr, args := selector.Query()
	expectedSQL := "SELECT * FROM `users` WHERE `name` = ?"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// 应该只有 name 条件
	if !strings.Contains(sqlStr, "name") {
		t.Error("Expected name condition in SQL")
	}
	if strings.Contains(sqlStr, "age") {
		t.Error("Expected age condition to be filtered out")
	}

	// 应该只有一个参数
	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
}

// 测试 OrderByHandler 选项
func TestWithOrderByHandler(t *testing.T) {
	// 创建 OrderBy 条件
	var orders clause.OrderBys
	orders = append(orders, clause.OrderBy{Column: "name", Desc: false})
	orders = append(orders, clause.OrderBy{Column: "age", Desc: true})

	// 创建一个排序处理器，修改列名
	handler := func(order clause.OrderBy) clause.OrderBy {
		if order.Column == "name" {
			order.Column = "user_name" // 修改列名
		}
		return order
	}

	// 转换为 ent OrderBy 函数
	orderByFunc := OrderBy(orders, WithOrderByHandler(handler))

	// 创建 ent Selector
	selector := sql.Select("*").From(sql.Table("users"))

	// 应用 OrderBy 函数
	orderByFunc(selector)

	// 检查生成的 SQL
	sqlStr, _ := selector.Query()
	expectedSQL := "SELECT * FROM `users` ORDER BY `user_name` ASC, `age` DESC"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// 应该包含修改后的列名
	if !strings.Contains(sqlStr, "user_name") {
		t.Error("Expected modified column name 'user_name' in SQL")
	}
}

// 测试复杂的嵌套条件
func TestComplexNestedConditions(t *testing.T) {
	// 创建复杂的查询条件: (name = 'John' AND age > 18) OR (city = 'New York' AND age < 65)
	q := query.Table("").
		OrWhere(func(w query.Wherer) query.Wherer {
			w.Where("name", "John")
			w.Where("age", ">", 18)
			return w
		}).
		OrWhere(func(w query.Wherer) query.Wherer {
			w.Where("city", "New York")
			w.Where("age", "<", 65)
			return w
		})

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
	expectedSQL := "SELECT * FROM `users` WHERE (`name` = ? AND `age` > ?) OR (`city` = ? AND `age` < ?)"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// 验证所有条件存在
	if !strings.Contains(sqlStr, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}
	if len(args) != 4 {
		t.Errorf("Expected 4 args, got %d", len(args))
	}
}

// 测试混合 AND、OR、NOT 条件
func TestMixedConditions(t *testing.T) {
	// 创建混合条件
	q := query.Table("").
		Where("city", "New York").
		OrWhere(func(w query.Wherer) query.Wherer {
			w.Where("name", "John")
			w.Not("age", ">", 65)
			return w
		})

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
	expectedSQL := "SELECT * FROM `users` WHERE `city` = ? OR (`name` = ? AND (NOT (`age` > ?)))"
	if sqlStr != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sqlStr)
	}

	// 验证混合条件存在
	if !strings.Contains(sqlStr, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}
	if !strings.Contains(sqlStr, "NOT") {
		t.Error("Expected NOT in SQL")
	}
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}
}

// 测试 LIKE 值必须是字符串的错误处理
func TestLikeWhereFunc_InvalidValue(t *testing.T) {
	// 创建一个错误的 LIKE 条件（值不是字符串）
	where := clause.Where{
		Exprs: []clause.Expression{
			clause.Like{Column: "name", Value: 123}, // 错误：值不是字符串
		},
	}

	// 转换为 ent Where 函数
	whereFunc := Where(where)

	// 创建 ent Selector
	selector := sql.Select("*").From(sql.Table("users"))

	// 应用 Where 函数
	whereFunc(selector)

	// 检查生成的 SQL - 由于错误，SQL 应该受到影响
	selector.Query()

	err := selector.Err()
	if err == nil {
		t.Error("Expected error")
	}
}

// Benchmark 测试
func BenchmarkWhereFunc(b *testing.B) {
	q := query.Where("name", "John").Where("age", ">", 18)
	where := q.WhereExpr()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		whereFunc := Where(where)
		selector := sql.Select("*").From(sql.Table("users"))
		whereFunc(selector)
		selector.Query()
	}
}

func BenchmarkQuery(b *testing.B) {
	q := query.Where("name", "John").Where("age", ">", 18)
	where := q.WhereExpr()

	var orders clause.OrderBys
	orders = append(orders, clause.OrderBy{Column: "name", Desc: false})

	limit := 10
	pagination := clause.Pagination{Limit: &limit, Offset: 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queryFunc := Query(where, orders, pagination)
		selector := sql.Select("*").From(sql.Table("users"))
		queryFunc(selector)
		selector.Query()
	}
}

// 测试示例：完整的使用场景
func ExampleWhere() {
	// 创建查询条件
	q := query.Where("name", "John").Where("age", ">", 18)
	where := q.WhereExpr()

	// 转换为 ent Where 函数
	whereFunc := Where(where)

	// 在实际使用中应用
	selector := sql.Select("*").From(sql.Table("users"))
	whereFunc(selector)

	sqlStr, _ := selector.Query()
	fmt.Printf("Generated SQL with WHERE clause: %v", strings.Contains(sqlStr, "WHERE"))
	// Output: Generated SQL with WHERE clause: true
}
