package aip

import (
	"fmt"
	"strings"
	"testing"

	"github.com/epkgs/query/clause"
	filtering "go.einride.tech/aip/filtering"
)

// mockFilterRequest 是一个用于测试的 Request 实现
// 它实现了 filtering.Request 接口，返回固定的 Filter 字符串
type mockFilterRequest struct {
	filter string
}

// GetFilter 返回 Filter 字符串
func (r mockFilterRequest) GetFilter() string {
	return r.filter
}

// mockBuilder 是一个用于测试的 Builder 实现
// 它捕获构建的 SQL 字符串和参数
type mockBuilder struct {
	sql  string
	vars []interface{}
	err  error
}

// WriteString 写入字符串到 SQL，实现 Writer 接口
func (m *mockBuilder) WriteString(s string) (int, error) {
	m.sql += s
	return len(s), nil
}

// WriteByte 写入字节到 SQL，实现 Writer 接口
func (m *mockBuilder) WriteByte(b byte) error {
	m.sql += string(b)
	return nil
}

// WriteQuoted 写入带引号的字符串到 SQL，实现 Builder 接口
func (m *mockBuilder) WriteQuoted(field interface{}) {
	if s, ok := field.(string); ok {
		m.sql += "`" + s + "`"
	} else {
		m.sql += "`" + "unknown" + "`"
	}
}

// AddVar 添加参数到变量列表，实现 Builder 接口
func (m *mockBuilder) AddVar(writer clause.Writer, vars ...interface{}) {
	for i, v := range vars {
		m.vars = append(m.vars, v)

		if i > 0 {
			m.sql += " "
		}
		m.sql += "$" + fmt.Sprintf("%d", len(m.vars))
	}
}

// AddError 添加错误，实现 Builder 接口
func (m *mockBuilder) AddError(err error) error {
	if m.err == nil {
		m.err = err
	}
	return m.err
}

// String 返回生成的 SQL
func (m *mockBuilder) String() string {
	return m.sql
}

// GetError 返回捕获的错误
func (m *mockBuilder) GetError() error {
	return m.err
}

// 辅助函数：测试Filter转换
func testFilterConversion(t *testing.T, filterStr, expectedSQL string, expectedArgs int) {
	// 创建一个声明集合
	declarations, err := filtering.NewDeclarations(
		filtering.DeclareIdent("name", filtering.TypeString),
		filtering.DeclareIdent("age", filtering.TypeInt),
		filtering.DeclareIdent("email", filtering.TypeString),
		filtering.DeclareIdent("status", filtering.TypeString),
		filtering.DeclareStandardFunctions(),
	)
	if err != nil {
		t.Fatalf("Failed to create declarations: %v", err)
	}

	// 创建一个测试请求
	req := mockFilterRequest{
		filter: filterStr,
	}

	// 解析 Filter
	filter, err := filtering.ParseFilter(req, declarations)
	if err != nil {
		t.Fatalf("Failed to parse filter: %v", err)
	}

	// 调用 FromFilter 将 Filter 转换为 SelectQuery
	w, err := FromFilter(filter)
	if err != nil {
		t.Fatalf("Failed to convert filter: %v", err)
	}

	// 创建一个简单的 mockBuilder 来验证 SQL 生成
	builder := &mockBuilder{}

	// 构建 SQL
	w.Build(builder)

	got := strings.TrimPrefix(builder.String(), " WHERE ")

	// 验证生成的 SQL 是否包含预期的条件
	if got != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, got)
	}

	// 验证参数数量
	if len(builder.vars) != expectedArgs {
		t.Errorf("Expected %d args, got %d", expectedArgs, len(builder.vars))
	}
}

// TestFromFilter_Integration 测试完整的集成功能
func TestFromFilter_Integration(t *testing.T) {
	testFilterConversion(t, "name = 'John' AND age > 18",
		"`name` = $1 AND `age` > $2", 2)
}

// TestFromFilter_ORCondition 测试 OR 条件
func TestFromFilter_ORCondition(t *testing.T) {
	testFilterConversion(
		t,
		"name = 'John' OR name = 'Jane'",
		"`name` = $1 OR `name` = $2",
		2,
	)
}

// TestFromFilter_NotCondition 测试 NOT 条件
func TestFromFilter_NotCondition(t *testing.T) {
	testFilterConversion(
		t,
		"NOT age < 18",
		"`age` >= $1",
		1,
	)
}

// TestFromFilter_CombinedCondition 测试组合条件
func TestFromFilter_CombinedCondition(t *testing.T) {
	testFilterConversion(
		t,
		"(name = 'John' OR name = 'Jane') AND age > 18",
		"(`name` = $1 OR `name` = $2) AND `age` > $3",
		3,
	)
}

// TestFromFilter_ComparisonOperators 测试比较运算符
func TestFromFilter_ComparisonOperators(t *testing.T) {
	testFilterConversion(
		t,
		"age >= 18 AND age <= 30",
		"`age` >= $1 AND `age` <= $2",
		2,
	)
}

// TestFromFilter_MultiFieldCondition 测试多字段条件
func TestFromFilter_MultiFieldCondition(t *testing.T) {
	testFilterConversion(
		t,
		"name = 'John' AND age > 18 AND status = 'active'",
		"(`name` = $1 AND `age` > $2) AND `status` = $3",
		3,
	)
}

// TestFromFilter_EmptyFilter 测试空Filter
func TestFromFilter_EmptyFilter(t *testing.T) {
	testFilterConversion(
		t,
		"",
		"",
		0,
	)
}

// TestFromFilter_INCondition 测试 IN 条件
// 注意：AIP Filter 可能不支持直接的 IN 语法，使用 OR 条件替代
func TestFromFilter_INCondition(t *testing.T) {
	testFilterConversion(
		t,
		"name = 'John' OR name = 'Jane' OR name = 'Bob'",
		"(`name` = $1 OR `name` = $2) OR `name` = $3",
		3,
	)
}

// TestFromFilter_ComplexNestedCondition 测试复杂嵌套条件
// 使用类型兼容的条件，避免 OR 函数的类型不匹配问题
func TestFromFilter_ComplexNestedCondition(t *testing.T) {
	testFilterConversion(
		t,
		"(name = 'John' AND (age > 18 OR age < 30)) OR (name = 'Jane' AND age < 30)",
		"(`name` = $1 AND (`age` > $2 OR `age` < $3)) OR (`name` = $4 AND `age` < $5)",
		5,
	)
}
