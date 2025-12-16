package aip

import (
	"testing"

	ordering "go.einride.tech/aip/ordering"
)

// TestFromOrderBy_SingleFieldAsc 测试单个字段升序排序
func TestFromOrderBy_SingleFieldAsc(t *testing.T) {
	// 创建测试用的 ordering.OrderBy
	aipOrderBy := ordering.OrderBy{
		Fields: []ordering.Field{
			{Path: "name", Desc: false},
		},
	}

	// 调用 FromOrderBy 转换
	clauseOrderBys := FromOrderBy(aipOrderBy)

	// 验证结果
	if clauseOrderBys == nil {
		t.Fatal("Expected non-nil *clause.OrderBys")
	}

	if len(clauseOrderBys) != 1 {
		t.Errorf("Expected 1 orderby, got %d", len(clauseOrderBys))
		return
	}

	orderby := (clauseOrderBys)[0]
	if orderby.Column != "name" {
		t.Errorf("Expected Column 'name', got '%s'", orderby.Column)
	}

	if orderby.Desc != false {
		t.Errorf("Expected Desc false, got %v", orderby.Desc)
	}
}

// TestFromOrderBy_SingleFieldDesc 测试单个字段降序排序
func TestFromOrderBy_SingleFieldDesc(t *testing.T) {
	// 创建测试用的 ordering.OrderBy
	aipOrderBy := ordering.OrderBy{
		Fields: []ordering.Field{
			{Path: "age", Desc: true},
		},
	}

	// 调用 FromOrderBy 转换
	clauseOrderBys := FromOrderBy(aipOrderBy)

	// 验证结果
	if clauseOrderBys == nil {
		t.Fatal("Expected non-nil *clause.OrderBys")
	}

	if len(clauseOrderBys) != 1 {
		t.Errorf("Expected 1 orderby, got %d", len(clauseOrderBys))
		return
	}

	orderby := (clauseOrderBys)[0]
	if orderby.Column != "age" {
		t.Errorf("Expected Column 'age', got '%s'", orderby.Column)
	}

	if orderby.Desc != true {
		t.Errorf("Expected Desc true, got %v", orderby.Desc)
	}
}

// TestFromOrderBy_MultipleFields 测试多个字段排序
func TestFromOrderBy_MultipleFields(t *testing.T) {
	// 创建测试用的 ordering.OrderBy
	aipOrderBy := ordering.OrderBy{
		Fields: []ordering.Field{
			{Path: "name", Desc: false},
			{Path: "age", Desc: true},
			{Path: "created_at", Desc: true},
		},
	}

	// 调用 FromOrderBy 转换
	clauseOrderBys := FromOrderBy(aipOrderBy)

	// 验证结果
	if clauseOrderBys == nil {
		t.Fatal("Expected non-nil *clause.OrderBys")
	}

	if len(clauseOrderBys) != 3 {
		t.Errorf("Expected 3 orderbys, got %d", len(clauseOrderBys))
		return
	}

	// 验证第一个字段
	orderby1 := (clauseOrderBys)[0]
	if orderby1.Column != "name" || orderby1.Desc != false {
		t.Errorf("Expected orderby 0: Column 'name', Desc false, got Column '%s', Desc %v", orderby1.Column, orderby1.Desc)
	}

	// 验证第二个字段
	orderby2 := (clauseOrderBys)[1]
	if orderby2.Column != "age" || orderby2.Desc != true {
		t.Errorf("Expected orderby 1: Column 'age', Desc true, got Column '%s', Desc %v", orderby2.Column, orderby2.Desc)
	}

	// 验证第三个字段
	orderby3 := (clauseOrderBys)[2]
	if orderby3.Column != "created_at" || orderby3.Desc != true {
		t.Errorf("Expected orderby 2: Column 'created_at', Desc true, got Column '%s', Desc %v", orderby3.Column, orderby3.Desc)
	}
}

// TestFromOrderBy_Empty 测试空排序条件
func TestFromOrderBy_Empty(t *testing.T) {
	// 创建测试用的空 ordering.OrderBy
	aipOrderBy := ordering.OrderBy{
		Fields: []ordering.Field{},
	}

	// 调用 FromOrderBy 转换
	clauseOrderBys := FromOrderBy(aipOrderBy)

	// 验证结果
	if clauseOrderBys == nil {
		t.Fatal("Expected non-nil *clause.OrderBys")
	}

	if len(clauseOrderBys) != 0 {
		t.Errorf("Expected 0 orderbys, got %d", len(clauseOrderBys))
	}
}

// TestFromOrderBy_ComplexPath 测试复杂路径排序
func TestFromOrderBy_ComplexPath(t *testing.T) {
	// 创建测试用的 ordering.OrderBy，包含复杂路径
	aipOrderBy := ordering.OrderBy{
		Fields: []ordering.Field{
			{Path: "user.name", Desc: false},
			{Path: "user.age", Desc: true},
		},
	}

	// 调用 FromOrderBy 转换
	clauseOrderBys := FromOrderBy(aipOrderBy)

	// 验证结果
	if clauseOrderBys == nil {
		t.Fatal("Expected non-nil *clause.OrderBys")
	}

	if len(clauseOrderBys) != 2 {
		t.Errorf("Expected 2 orderbys, got %d", len(clauseOrderBys))
		return
	}

	// 验证第一个复杂路径字段
	orderby1 := (clauseOrderBys)[0]
	if orderby1.Column != "user.name" || orderby1.Desc != false {
		t.Errorf("Expected orderby 0: Column 'user.name', Desc false, got Column '%s', Desc %v", orderby1.Column, orderby1.Desc)
	}

	// 验证第二个复杂路径字段
	orderby2 := (clauseOrderBys)[1]
	if orderby2.Column != "user.age" || orderby2.Desc != true {
		t.Errorf("Expected orderby 1: Column 'user.age', Desc true, got Column '%s', Desc %v", orderby2.Column, orderby2.Desc)
	}
}
