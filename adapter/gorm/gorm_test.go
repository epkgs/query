package gorm

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/epkgs/query"
	"github.com/epkgs/query/clause"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// User 测试用的模型
type User struct {
	ID   int
	Name string
	Age  int
	City string
}

// getTestDB 创建一个测试用的 DB 实例
func getTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	return db
}

// 测试基本的 Where 条件转换
func TestWhereScope(t *testing.T) {
	// 创建查询条件
	q := query.Where("name", "John").Where("age", ">", 18)

	// 获取 Where 表达式
	where := q.WhereExpr()

	// 转换为 gorm scope
	scope := Where(where)

	// 创建 gorm DB 实例
	db := getTestDB(t)

	// 应用 scope 并构建查询
	stmt := db.Model(&User{}).Scopes(scope).Find(&[]User{}).Statement

	// 验证生成的 SQL
	sql := stmt.SQL.String()
	vars := stmt.Vars

	t.Logf("SQL: %s", sql)
	t.Logf("Vars: %v", vars)

	// 检查 SQL 包含预期的条件
	if !strings.Contains(sql, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}
	if !strings.Contains(sql, "`name` = ?") {
		t.Error("Expected name condition in SQL")
	}
	if !strings.Contains(sql, "`age` > ?") {
		t.Error("Expected age condition in SQL")
	}

	// 检查变量
	if len(vars) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(vars))
		return
	}
	if vars[0] != "John" {
		t.Errorf("Expected first var to be 'John', got %v", vars[0])
	}
	if vars[1] != 18 {
		t.Errorf("Expected second var to be 18, got %v", vars[1])
	}
}

// 测试空 Where 条件
func TestWhereScope_Empty(t *testing.T) {
	// 创建空的查询条件
	where := clause.Where{}

	// 转换为 gorm scope
	scope := Where(where)

	// 创建 gorm DB 实例
	db := getTestDB(t)

	// 应用 scope 并构建查询
	stmt := db.Model(&User{}).Scopes(scope).Find(&[]User{}).Statement

	// 验证生成的 SQL
	sql := stmt.SQL.String()

	t.Logf("SQL: %s", sql)

	// 空条件不应该有 WHERE 子句
	if strings.Contains(sql, "WHERE") {
		t.Error("Expected no WHERE clause for empty conditions")
	}
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
	db := getTestDB(t)

	// 应用 scope 并构建查询
	stmt := db.Model(&User{}).Scopes(scope).Find(&[]User{}).Statement

	// 验证生成的 SQL
	sql := stmt.SQL.String()

	t.Logf("SQL: %s", sql)

	// 检查 ORDER BY 子句
	if !strings.Contains(sql, "ORDER BY") {
		t.Error("Expected ORDER BY clause in SQL")
	}
	if !strings.Contains(sql, "`name`") {
		t.Error("Expected name in ORDER BY clause")
	}
	if !strings.Contains(sql, "`age` DESC") {
		t.Error("Expected age DESC in ORDER BY clause")
	}
}

// 测试空 OrderBy 条件
func TestOrderByScope_Empty(t *testing.T) {
	// 创建空的 OrderBy 条件
	var orders clause.OrderBys

	// 转换为 gorm scope
	scope := OrderBy(orders)

	// 创建 gorm DB 实例
	db := getTestDB(t)

	// 应用 scope 并构建查询
	stmt := db.Model(&User{}).Scopes(scope).Find(&[]User{}).Statement

	// 验证生成的 SQL
	sql := stmt.SQL.String()

	t.Logf("SQL: %s", sql)

	// 空条件不应该有 ORDER BY 子句
	if strings.Contains(sql, "ORDER BY") {
		t.Error("Expected no ORDER BY clause for empty conditions")
	}
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
	db := getTestDB(t)

	// 应用 scope 并构建查询
	stmt := db.Model(&User{}).Scopes(scope).Find(&[]User{}).Statement

	// 验证生成的 SQL
	sql := stmt.SQL.String()

	t.Logf("SQL: %s", sql)

	// 检查 LIMIT 和 OFFSET
	if !strings.Contains(sql, "LIMIT 10") {
		t.Error("Expected LIMIT 10 in SQL")
	}
	if !strings.Contains(sql, "OFFSET 20") {
		t.Error("Expected OFFSET 20 in SQL")
	}
}

// 测试只有 Limit 的 Pagination
func TestPaginationScope_LimitOnly(t *testing.T) {
	// 创建 Pagination 条件
	limit := 5
	pagination := clause.Pagination{
		Limit:  &limit,
		Offset: 0,
	}

	// 转换为 gorm scope
	scope := Pagination(pagination)

	// 创建 gorm DB 实例
	db := getTestDB(t)

	// 应用 scope 并构建查询
	stmt := db.Model(&User{}).Scopes(scope).Find(&[]User{}).Statement

	// 验证生成的 SQL
	sql := stmt.SQL.String()

	t.Logf("SQL: %s", sql)

	// 检查 LIMIT
	if !strings.Contains(sql, "LIMIT 5") {
		t.Error("Expected LIMIT 5 in SQL")
	}
}

// 测试完整的 Query 条件转换
func TestQueryScope(t *testing.T) {
	// 创建 Where 条件
	q := query.Where("name", "John").Where("age", ">", 18)
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
	db := getTestDB(t)

	// 应用 scope 并构建查询
	stmt := db.Model(&User{}).Scopes(scope).Find(&[]User{}).Statement

	// 验证生成的 SQL
	sql := stmt.SQL.String()
	vars := stmt.Vars

	t.Logf("SQL: %s", sql)
	t.Logf("Vars: %v", vars)

	// 检查所有组件都存在
	if !strings.Contains(sql, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}
	if !strings.Contains(sql, "ORDER BY") {
		t.Error("Expected ORDER BY clause in SQL")
	}
	if !strings.Contains(sql, "LIMIT 10") {
		t.Error("Expected LIMIT in SQL")
	}
	if !strings.Contains(sql, "OFFSET 20") {
		t.Error("Expected OFFSET in SQL")
	}

	// 检查变量
	if len(vars) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(vars))
	}
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
	db := getTestDB(t)

	// 应用 scope 并构建查询
	stmt := db.Model(&User{}).Scopes(scope).Find(&[]User{}).Statement

	// 验证生成的 SQL
	sql := stmt.SQL.String()
	vars := stmt.Vars

	t.Logf("SQL: %s", sql)
	t.Logf("Vars: %v", vars)

	// 检查 OR 条件
	if !strings.Contains(sql, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}
	if !strings.Contains(sql, "OR") {
		t.Error("Expected OR in SQL")
	}

	// 检查变量
	if len(vars) != 3 {
		t.Errorf("Expected 3 variables, got %d", len(vars))
	}
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
	db := getTestDB(t)

	// 应用 scope 并构建查询
	stmt := db.Model(&User{}).Scopes(scope).Find(&[]User{}).Statement

	// 验证生成的 SQL
	sql := stmt.SQL.String()
	vars := stmt.Vars

	t.Logf("SQL: %s", sql)
	t.Logf("Vars: %v", vars)

	// 检查 NOT 条件 - GORM 会将 NOT 转换为反向操作符
	// NOT (name = "John") 变为 name <> "John"
	// NOT (age > 18) 变为 age <= 18
	if !strings.Contains(sql, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}
	if !strings.Contains(sql, "`name` <>") {
		t.Error("Expected name <> (NOT equal) in SQL")
	}
	if !strings.Contains(sql, "`age` <=") {
		t.Error("Expected age <= (NOT greater than) in SQL")
	}

	// 检查变量
	if len(vars) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(vars))
	}
}

// 测试 IN 条件转换
func TestInWhereScope(t *testing.T) {
	// 创建查询条件
	q := query.Where("id", "IN", []interface{}{1, 2, 3})

	// 获取 Where 表达式
	where := q.WhereExpr()

	// 转换为 gorm scope
	scope := Where(where)

	// 创建 gorm DB 实例
	db := getTestDB(t)

	// 应用 scope 并构建查询
	stmt := db.Model(&User{}).Scopes(scope).Find(&[]User{}).Statement

	// 验证生成的 SQL
	sql := stmt.SQL.String()
	vars := stmt.Vars

	t.Logf("SQL: %s", sql)
	t.Logf("Vars: %v", vars)

	// 检查 IN 条件
	if !strings.Contains(sql, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}
	if !strings.Contains(sql, "`id` IN (?,?,?)") {
		t.Error("Expected IN clause in SQL")
	}

	// 检查变量
	if len(vars) != 3 {
		t.Errorf("Expected 3 variables, got %d", len(vars))
	}
	expectedVars := []interface{}{1, 2, 3}
	if !reflect.DeepEqual(vars, expectedVars) {
		t.Errorf("Expected vars %v, got %v", expectedVars, vars)
	}
}

// 测试 LIKE 条件转换
func TestLikeWhereScope(t *testing.T) {
	// 创建查询条件
	q := query.Where("name", "LIKE", "%John%")

	// 获取 Where 表达式
	where := q.WhereExpr()

	// 转换为 gorm scope
	scope := Where(where)

	// 创建 gorm DB 实例
	db := getTestDB(t)

	// 应用 scope 并构建查询
	stmt := db.Model(&User{}).Scopes(scope).Find(&[]User{}).Statement

	// 验证生成的 SQL
	sql := stmt.SQL.String()
	vars := stmt.Vars

	t.Logf("SQL: %s", sql)
	t.Logf("Vars: %v", vars)

	// 检查 LIKE 条件
	if !strings.Contains(sql, "LIKE") {
		t.Error("Expected LIKE in SQL")
	}

	// 检查变量
	if len(vars) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(vars))
		return
	}
	if vars[0] != "%John%" {
		t.Errorf("Expected var to be '%%John%%', got %v", vars[0])
	}
}

// 测试所有比较操作符
func TestComparisonOperators(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		value    interface{}
		sqlPart  string
	}{
		{"Equal", "=", "John", "`name` = ?"},
		{"NotEqual", "!=", "John", "`name` <> ?"},
		{"GreaterThan", ">", 18, "`age` > ?"},
		{"GreaterOrEqual", ">=", 18, "`age` >= ?"},
		{"LessThan", "<", 65, "`age` < ?"},
		{"LessOrEqual", "<=", 65, "`age` <= ?"},
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

			// 转换为 gorm scope
			scope := Where(where)

			// 创建 gorm DB 实例
			db := getTestDB(t)

			// 应用 scope 并构建查询
			stmt := db.Model(&User{}).Scopes(scope).Find(&[]User{}).Statement

			// 验证生成的 SQL
			sql := stmt.SQL.String()

			t.Logf("SQL: %s", sql)

			if !strings.Contains(sql, tt.sqlPart) {
				t.Errorf("Expected SQL to contain '%s', got: %s", tt.sqlPart, sql)
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

	// 转换为 gorm scope
	scope := Where(where, WithExprHandler(handler))

	// 创建 gorm DB 实例
	db := getTestDB(t)

	// 应用 scope 并构建查询
	stmt := db.Model(&User{}).Scopes(scope).Find(&[]User{}).Statement

	// 验证生成的 SQL
	sql := stmt.SQL.String()
	vars := stmt.Vars

	t.Logf("SQL: %s", sql)
	t.Logf("Vars: %v", vars)

	// 应该只有 name 条件
	if !strings.Contains(sql, "`name` = ?") {
		t.Error("Expected name condition in SQL")
	}
	if strings.Contains(sql, "`age`") {
		t.Error("Expected age condition to be filtered out")
	}

	// 应该只有一个变量
	if len(vars) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(vars))
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

	// 转换为 gorm scope
	scope := OrderBy(orders, WithOrderByHandler(handler))

	// 创建 gorm DB 实例
	db := getTestDB(t)

	// 应用 scope 并构建查询
	stmt := db.Model(&User{}).Scopes(scope).Find(&[]User{}).Statement

	// 验证生成的 SQL
	sql := stmt.SQL.String()

	t.Logf("SQL: %s", sql)

	// 应该包含修改后的列名
	if !strings.Contains(sql, "`user_name`") {
		t.Error("Expected modified column name 'user_name' in SQL")
	}
	// 在 ORDER BY 中不应该出现原始的 name（但可能在 SELECT 中出现）
	orderByPart := sql[strings.Index(sql, "ORDER BY"):]
	if strings.Contains(orderByPart, "`name`") && !strings.Contains(orderByPart, "`user_name`") {
		t.Error("Expected original column name 'name' to be replaced in ORDER BY")
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

	// 转换为 gorm scope
	scope := Where(where)

	// 创建 gorm DB 实例
	db := getTestDB(t)

	// 应用 scope 并构建查询
	stmt := db.Model(&User{}).Scopes(scope).Find(&[]User{}).Statement

	// 验证生成的 SQL
	sql := stmt.SQL.String()
	vars := stmt.Vars

	t.Logf("SQL: %s", sql)
	t.Logf("Vars: %v", vars)

	// 检查所有条件
	if !strings.Contains(sql, "OR") {
		t.Error("Expected OR in SQL")
	}
	if len(vars) != 4 {
		t.Errorf("Expected 4 variables, got %d", len(vars))
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

	// 转换为 gorm scope
	scope := Where(where)

	// 创建 gorm DB 实例
	db := getTestDB(t)

	// 应用 scope 并构建查询
	stmt := db.Model(&User{}).Scopes(scope).Find(&[]User{}).Statement

	// 验证生成的 SQL
	sql := stmt.SQL.String()
	vars := stmt.Vars

	t.Logf("SQL: %s", sql)
	t.Logf("Vars: %v", vars)

	// 检查混合条件
	if !strings.Contains(sql, "WHERE") {
		t.Error("Expected WHERE clause in SQL")
	}
	if !strings.Contains(sql, "OR") {
		t.Error("Expected OR in SQL")
	}
	// NOT 条件会被 GORM 转换为反向操作符 (age > 65 变为 age <= 65)
	if !strings.Contains(sql, "<=") {
		t.Error("Expected <= (NOT greater than) in SQL")
	}
}

// Benchmark 测试
func BenchmarkWhereScope(b *testing.B) {
	q := query.Where("name", "John").Where("age", ">", 18)
	where := q.WhereExpr()

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scope := Where(where)
		scope(db)
	}
}

func BenchmarkQueryScope(b *testing.B) {
	q := query.Where("name", "John").Where("age", ">", 18)
	where := q.WhereExpr()

	var orders clause.OrderBys
	orders = append(orders, clause.OrderBy{Column: "name", Desc: false})

	limit := 10
	pagination := clause.Pagination{Limit: &limit, Offset: 20}

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scope := Query(where, orders, pagination)
		scope(db)
	}
}

// 测试示例：完整的使用场景
func ExampleWhere() {
	// 创建查询条件
	q := query.Where("name", "John").Where("age", ">", 18)
	where := q.WhereExpr()

	// 转换为 gorm scope
	scope := Where(where)

	// 在实际使用中应用
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		DryRun: true,
	})
	var users []User
	db = db.Model(&User{}).Scopes(scope).Find(&users)

	fmt.Printf("SQL generated successfully")
	// Output: SQL generated successfully
}

