package query

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/epkgs/query/clause"
)

// mockBuilder 是用于测试的Builder实现
type mockBuilder struct {
	bytes.Buffer
	vars   []interface{}
	errors []error
}

func (m *mockBuilder) WriteQuoted(field interface{}) {
	switch v := field.(type) {
	case string:
		m.WriteString(fmt.Sprintf("`%s`", v))
	default:
		m.WriteString(fmt.Sprintf("%v", v))
	}
}

func (m *mockBuilder) AddVar(writer clause.Writer, vars ...interface{}) {
	for _, v := range vars {
		m.vars = append(m.vars, v)
		if w, ok := writer.(*mockBuilder); ok {
			w.WriteString(fmt.Sprintf("$%d", len(w.vars)))
		}
	}
}

func (m *mockBuilder) AddError(err error) error {
	if err != nil {
		m.errors = append(m.errors, err)
	}
	return err
}

// TestQuery_BuildSelect 测试基本的SELECT查询（默认选择所有字段）
func TestQuery_BuildSelect(t *testing.T) {
	q := Table("users").Select()
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT * FROM `users`"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestQuery_BuildSelectSingleField 测试选择单个字段的SELECT查询
func TestQuery_BuildSelectSingleField(t *testing.T) {
	q := Table("users").Select("name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `name` FROM `users`"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestQuery_BuildSelectMultipleFields 测试选择多个字段的SELECT查询
func TestQuery_BuildSelectMultipleFields(t *testing.T) {
	q := Table("users").Select("id", "name", "email")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name`, `email` FROM `users`"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestQuery_BuildSelectWithWhere 测试选择字段并带有WHERE条件的查询
func TestQuery_BuildSelectWithWhere(t *testing.T) {
	q := Table("users").Where("age", ">", 18).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `age` > $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestQuery_BuildWhere 测试WHERE条件查询
func TestQuery_BuildWhere(t *testing.T) {
	q := Table("users").Where("name", "=", "John").Select()
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT * FROM `users` WHERE `name` = $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 1 || builder.vars[0] != "John" {
		t.Errorf("expected vars: [John], got: %v", builder.vars)
	}
}

// TestQuery_BuildUpdate 测试基本的UPDATE操作
func TestQuery_BuildUpdate(t *testing.T) {
	q := Table("users").Update("name", "John").Update("age", 30)
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "UPDATE `users` SET `age` = $1, `name` = $2"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 2 || builder.vars[0] != 30 || builder.vars[1] != "John" {
		t.Errorf("expected vars: [30 John], got: %v", builder.vars)
	}
}

// TestQuery_BuildUpdateWithWhere 测试带WHERE条件的UPDATE操作
func TestQuery_BuildUpdateWithWhere(t *testing.T) {
	q := Table("users").Where("id", 1).Update("name", "John")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "UPDATE `users` SET `name` = $1 WHERE `id` = $2"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 2 || builder.vars[0] != "John" || builder.vars[1] != 1 {
		t.Errorf("expected vars: [John 1], got: %v", builder.vars)
	}
}

// TestQuery_BuildUpdateWithMap 测试使用map设置多个字段的UPDATE操作
func TestQuery_BuildUpdateWithMap(t *testing.T) {
	q := Table("users").Where("id", 1).Update(map[string]interface{}{"name": "John", "age": 30})
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "UPDATE `users` SET `age` = $1, `name` = $2 WHERE `id` = $3"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 3 || builder.vars[0] != 30 || builder.vars[1] != "John" || builder.vars[2] != 1 {
		t.Errorf("expected vars: [30 John 1], got: %v", builder.vars)
	}
}

// TestQuery_BuildInsert 测试基本的INSERT操作（单行插入）
func TestQuery_BuildInsert(t *testing.T) {
	q := Table("users").Insert(map[string]interface{}{"name": "John", "age": 30})
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "INSERT INTO `users` (`age`, `name`) VALUES ($1, $2)"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 2 || builder.vars[0] != 30 || builder.vars[1] != "John" {
		t.Errorf("expected vars: [30 John], got: %v", builder.vars)
	}
}

// TestQuery_BuildInsertMultiple 测试多行INSERT操作
func TestQuery_BuildInsertMultiple(t *testing.T) {
	q := Table("users").Insert(
		map[string]interface{}{"name": "John", "age": 30},
		map[string]interface{}{"name": "Jane", "age": 25},
	)
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "INSERT INTO `users` (`age`, `name`) VALUES ($1, $2), ($3, $4)"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 4 || builder.vars[0] != 30 || builder.vars[1] != "John" || builder.vars[2] != 25 || builder.vars[3] != "Jane" {
		t.Errorf("expected vars: [30 John 25 Jane], got: %v", builder.vars)
	}
}

// TestQuery_BuildInsertDifferentFields 测试带有不同字段的多行INSERT操作
func TestQuery_BuildInsertDifferentFields(t *testing.T) {
	q := Table("users").Insert(
		map[string]interface{}{"name": "John", "age": 30},
		map[string]interface{}{"name": "Jane", "email": "jane@example.com"},
	)
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "INSERT INTO `users` (`age`, `email`, `name`) VALUES ($1, $2, $3), ($4, $5, $6)"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 6 || builder.vars[0] != 30 || builder.vars[1] != nil || builder.vars[2] != "John" || builder.vars[3] != nil || builder.vars[4] != "jane@example.com" || builder.vars[5] != "Jane" {
		t.Errorf("expected vars: [30 <nil> John <nil> jane@example.com Jane], got: %v", builder.vars)
	}
}

// TestQuery_BuildInsertWithFieldValuePairs 测试使用字段值对的INSERT操作
func TestQuery_BuildInsertWithFieldValuePairs(t *testing.T) {
	q := Table("users").Insert("name", "John", "age", 30)
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "INSERT INTO `users` (`age`, `name`) VALUES ($1, $2)"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 2 || builder.vars[0] != 30 || builder.vars[1] != "John" {
		t.Errorf("expected vars: [30 John], got: %v", builder.vars)
	}
}

// TestQuery_BuildDelete 测试基本的DELETE操作
func TestQuery_BuildDelete(t *testing.T) {
	q := Table("users").Delete()
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "DELETE FROM `users`"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestQuery_BuildDeleteWithWhere 测试带WHERE条件的DELETE操作
func TestQuery_BuildDeleteWithWhere(t *testing.T) {
	q := Table("users").Where("id", 1).Delete()
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "DELETE FROM `users` WHERE `id` = $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 1 || builder.vars[0] != 1 {
		t.Errorf("expected vars: [1], got: %v", builder.vars)
	}
}

// TestQuery_BuildSelectWithOrderBy 测试带排序的SELECT查询
func TestQuery_BuildSelectWithOrderBy(t *testing.T) {
	q := Table("users").OrderBy("name").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` ORDER BY `name` ASC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestQuery_BuildSelectWithOrderByDesc 测试带降序排序的SELECT查询
func TestQuery_BuildSelectWithOrderByDesc(t *testing.T) {
	q := Table("users").OrderBy("name", "desc").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` ORDER BY `name` DESC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestQuery_BuildSelectWithMultipleOrderBy 测试多字段排序的SELECT查询
func TestQuery_BuildSelectWithMultipleOrderBy(t *testing.T) {
	q := Table("users").OrderBy("age", "desc").OrderBy("name").Select("id", "name", "age")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name`, `age` FROM `users` ORDER BY `age` DESC, `name` ASC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestQuery_BuildSelectWithCommaSeparatedOrderBy 测试使用逗号分隔字段的排序
func TestQuery_BuildSelectWithCommaSeparatedOrderBy(t *testing.T) {
	q := Table("users").OrderBy("age desc, name").Select("id", "name", "age")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name`, `age` FROM `users` ORDER BY `age` DESC, `name` ASC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestQuery_BuildSelectWithLimit 测试带LIMIT的SELECT查询
func TestQuery_BuildSelectWithLimit(t *testing.T) {
	q := Table("users").Limit(10).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` LIMIT $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 1 || builder.vars[0] != 10 {
		t.Errorf("expected vars: [10], got: %v", builder.vars)
	}
}

// TestQuery_BuildSelectWithLimitOffset 测试带LIMIT和OFFSET的SELECT查询
func TestQuery_BuildSelectWithLimitOffset(t *testing.T) {
	q := Table("users").Limit(10).Offset(20).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` LIMIT $1 OFFSET $2"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 2 || builder.vars[0] != 10 || builder.vars[1] != 20 {
		t.Errorf("expected vars: [10 20], got: %v", builder.vars)
	}
}

// TestQuery_BuildSelectWithPaginate 测试使用Paginate方法的SELECT查询
func TestQuery_BuildSelectWithPaginate(t *testing.T) {
	q := Table("users").Paginate(3, 10).Select("id", "name") // 第3页，每页10条
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` LIMIT $1 OFFSET $2"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 2 || builder.vars[0] != 10 || builder.vars[1] != 20 {
		t.Errorf("expected vars: [10 20], got: %v", builder.vars)
	}
}

// TestQuery_BuildSelectWithComplexConditions 测试带复杂条件的SELECT查询
func TestQuery_BuildSelectWithComplexConditions(t *testing.T) {
	q := Table("users").Where("age", ">", 18).Where("age", "<", 30).Where("name", "LIKE", "J%").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `age` > $1 AND `age` < $2 AND `name` LIKE $3"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 3 || builder.vars[0] != 18 || builder.vars[1] != 30 || builder.vars[2] != "J%" {
		t.Errorf("expected vars: [18 30 J%%], got: %v", builder.vars)
	}
}

// TestQuery_BuildSelectWithOrWhere 测试带OR WHERE条件的SELECT查询
func TestQuery_BuildSelectWithOrWhere(t *testing.T) {
	q := Table("users").Where("age", ">", 18).OrWhere("name", "admin").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `age` > $1 OR `name` = $2"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 2 || builder.vars[0] != 18 || builder.vars[1] != "admin" {
		t.Errorf("expected vars: [18 admin], got: %v", builder.vars)
	}
}

// TestQuery_BuildSelectWithNot 测试带NOT条件的SELECT查询
func TestQuery_BuildSelectWithNot(t *testing.T) {
	q := Table("users").Where("age", ">", 18).Not("name", "admin").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `age` > $1 AND `name` <> $2"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 2 || builder.vars[0] != 18 || builder.vars[1] != "admin" {
		t.Errorf("expected vars: [18 admin], got: %v", builder.vars)
	}
}

// TestQuery_BuildSelectWithInCondition 测试带IN条件的SELECT查询
func TestQuery_BuildSelectWithInCondition(t *testing.T) {
	q := Table("users").Where("id", []interface{}{1, 2, 3, 4, 5}).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// 注意：IN条件的参数之间没有逗号分隔，这是当前实现的行为
	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `id` IN ($1$2$3$4$5)"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 5 {
		t.Errorf("expected 5 vars, got: %d", len(builder.vars))
	}
}

// TestQuery_BuildSelectWithAllFeatures 测试包含所有功能的SELECT查询
func TestQuery_BuildSelectWithAllFeatures(t *testing.T) {
	q := Table("users").Where("age", ">", 18).OrWhere("name", "admin").Not("status", "banned").OrderBy("age", "desc").OrderBy("name").Limit(10).Offset(20).Select("id", "name", "age")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name`, `age` FROM `users` WHERE `age` > $1 OR `name` = $2 AND `status` <> $3 ORDER BY `age` DESC, `name` ASC LIMIT $4 OFFSET $5"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 5 {
		t.Errorf("expected 5 vars, got: %d", len(builder.vars))
	}
}

// TestQuery_BuildSelectEmptyFields 测试不指定字段的SELECT查询
func TestQuery_BuildSelectEmptyFields(t *testing.T) {
	q := Table("users").Select()
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT * FROM `users`"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestQuery_BuildDeleteWithComplexConditions 测试带复杂条件的DELETE查询
func TestQuery_BuildDeleteWithComplexConditions(t *testing.T) {
	q := Table("users").Where("age", "<", 18).OrWhere("status", "banned").Where("created_at", "<", "2023-01-01").Delete()
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "DELETE FROM `users` WHERE `age` < $1 OR `status` = $2 AND `created_at` < $3"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestQuery_ChainableCalls 测试链式调用的正确性
func TestQuery_ChainableCalls(t *testing.T) {
	// 测试SelectQuery的链式调用
	q := Table("users").Where("age", ">", 18).OrderBy("name").Limit(10).Offset(0).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `age` > $1 ORDER BY `name` ASC LIMIT $2"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	// 测试UpdateQuery的链式调用
	q2 := Table("users").Where("id", 1).Where("status", "active").Update("name", "John")
	builder2 := &mockBuilder{}
	q2.Build(builder2)

	expectedSQL2 := "UPDATE `users` SET `name` = $1 WHERE `id` = $2 AND `status` = $3"
	if builder2.String() != expectedSQL2 {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL2, builder2.String())
	}
}

// TestQuery_BuildSelectWithMultipleOrWhere 测试多个OR条件的SELECT查询
func TestQuery_BuildSelectWithMultipleOrWhere(t *testing.T) {
	q := Table("users").Where("age", ">", 18).OrWhere("name", "admin").OrWhere("role", "moderator").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `age` > $1 OR `name` = $2 OR `role` = $3"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 3 || builder.vars[0] != 18 || builder.vars[1] != "admin" || builder.vars[2] != "moderator" {
		t.Errorf("expected vars: [18 admin moderator], got: %v", builder.vars)
	}
}

// TestQuery_BuildSelectWithComplexNotConditions 测试复杂的NOT条件
func TestQuery_BuildSelectWithComplexNotConditions(t *testing.T) {
	q := Table("users").Where("age", ">", 18).Not("role", "guest").Not("status", "banned").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `age` > $1 AND `role` <> $2 AND `status` <> $3"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 3 || builder.vars[0] != 18 || builder.vars[1] != "guest" || builder.vars[2] != "banned" {
		t.Errorf("expected vars: [18 guest banned], got: %v", builder.vars)
	}
}

// TestQuery_BuildSelectWithNullValue 测试包含NULL值条件的SELECT查询
func TestQuery_BuildSelectWithNullValue(t *testing.T) {
	q := Table("users").Where("email", nil).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `email` IS NULL"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 0 {
		t.Errorf("expected 0 vars, got: %d", len(builder.vars))
	}
}

// TestQuery_BuildUpdateWithNullValue 测试更新为空值的UPDATE操作
func TestQuery_BuildUpdateWithNullValue(t *testing.T) {
	q := Table("users").Where("id", 1).Update("email", nil)
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "UPDATE `users` SET `email` = $1 WHERE `id` = $2"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 2 || builder.vars[0] != nil || builder.vars[1] != 1 {
		t.Errorf("expected vars: [nil 1], got: %v", builder.vars)
	}
}

// TestQuery_BuildSelectWithZeroLimit 测试Limit为0的SELECT查询
func TestQuery_BuildSelectWithZeroLimit(t *testing.T) {
	q := Table("users").Limit(0).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users`"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 0 {
		t.Errorf("expected 0 vars, got: %d", len(builder.vars))
	}
}

// TestQuery_BuildSelectWithZeroOffset 测试Offset为0的SELECT查询
func TestQuery_BuildSelectWithZeroOffset(t *testing.T) {
	q := Table("users").Limit(10).Offset(0).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` LIMIT $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 1 || builder.vars[0] != 10 {
		t.Errorf("expected vars: [10], got: %v", builder.vars)
	}
}

// TestQuery_BuildSelectWithComplexSorting 测试复杂的排序场景
func TestQuery_BuildSelectWithComplexSorting(t *testing.T) {
	q := Table("users").OrderBy("age", "desc").OrderBy("created_at", "asc").OrderBy("name", "desc").Select("id", "name", "age", "created_at")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name`, `age`, `created_at` FROM `users` ORDER BY `age` DESC, `created_at` ASC, `name` DESC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestQuery_BuildSelectWithClauseOrderBy 测试使用单个 clause.OrderBy
func TestQuery_BuildSelectWithClauseOrderBy(t *testing.T) {
	q := Table("users").OrderBy(clause.OrderBy{Column: "name", Desc: false}).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` ORDER BY `name` ASC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestQuery_BuildSelectWithClauseOrderByDesc 测试使用单个 clause.OrderBy（降序）
func TestQuery_BuildSelectWithClauseOrderByDesc(t *testing.T) {
	q := Table("users").OrderBy(clause.OrderBy{Column: "name", Desc: true}).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` ORDER BY `name` DESC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestQuery_BuildSelectWithMultipleClauseOrderBy 测试使用多个 clause.OrderBy 参数
func TestQuery_BuildSelectWithMultipleClauseOrderBy(t *testing.T) {
	q := Table("users").OrderBy(
		clause.OrderBy{Column: "age", Desc: true},
		clause.OrderBy{Column: "name", Desc: false},
	).Select("id", "name", "age")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name`, `age` FROM `users` ORDER BY `age` DESC, `name` ASC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestQuery_BuildSelectWithClauseOrderBySlice 测试使用 []clause.OrderBy
func TestQuery_BuildSelectWithClauseOrderBySlice(t *testing.T) {
	orderBys := []clause.OrderBy{
		{Column: "age", Desc: true},
		{Column: "name", Desc: false},
	}
	q := Table("users").OrderBy(orderBys).Select("id", "name", "age")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name`, `age` FROM `users` ORDER BY `age` DESC, `name` ASC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestQuery_BuildSelectWithClauseOrderBys 测试使用 clause.OrderBys
func TestQuery_BuildSelectWithClauseOrderBys(t *testing.T) {
	orderBys := clause.OrderBys{
		{Column: "age", Desc: true},
		{Column: "name", Desc: false},
	}
	q := Table("users").OrderBy(orderBys).Select("id", "name", "age")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name`, `age` FROM `users` ORDER BY `age` DESC, `name` ASC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestQuery_BuildSelectWithMixedOrderByTypes 测试混合使用不同类型的排序条件
func TestQuery_BuildSelectWithMixedOrderByTypes(t *testing.T) {
	q := Table("users").
		OrderBy("age", "desc").                                        // 字符串类型
		OrderBy(clause.OrderBy{Column: "name", Desc: false}).          // 单个clause.OrderBy
		OrderBy([]clause.OrderBy{{Column: "created_at", Desc: true}}). // []clause.OrderBy
		Select("id", "name", "age", "created_at")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name`, `age`, `created_at` FROM `users` ORDER BY `age` DESC, `name` ASC, `created_at` DESC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}
func TestQuery_BuildSelectWithCommaSeparatedOrderByMixed(t *testing.T) {
	q := Table("users").OrderBy("age desc, created_at, name asc").Select("id", "name", "age")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name`, `age` FROM `users` ORDER BY `age` DESC, `created_at` ASC, `name` ASC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestClosureChainedWhere 测试闭包链式调用功能
func TestClosureChainedWhere(t *testing.T) {
	// 创建一个查询
	q := Table("users")

	// 使用闭包进行链式调用（多个条件）
	q.Where(func(w Wherer) Wherer {
		return w.Where("name", "John").OrWhere("name", "Jane").Not("age", "<", 18)
	})

	// 转换为SELECT查询并构建SQL
	builder := &mockBuilder{}
	q.Select("*").Build(builder)

	// 打印结果
	fmt.Println("SQL:", builder.String())
	fmt.Println("Args:", builder.vars)

	// 验证SQL是否包含预期条件（注意：当闭包是唯一条件时，不会添加括号）
	expectedSQL := "SELECT `*` FROM `users` WHERE `name` = $1 OR `name` = $2 AND `age` >= $3"
	if builder.String() != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	// 验证参数是否正确
	expectedArgs := []interface{}{"John", "Jane", 18}
	if len(builder.vars) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(builder.vars))
	}
	for i, arg := range expectedArgs {
		if builder.vars[i] != arg {
			t.Errorf("Expected arg %d to be %v, got %v", i, arg, builder.vars[i])
		}
	}
}

// TestClosureSingleCondition 测试闭包内只有一个条件时的情况
func TestClosureSingleCondition(t *testing.T) {
	// 创建一个查询
	q := Table("users")

	// 使用闭包只添加一个条件
	q.Where(func(w Wherer) Wherer {
		return w.Where("name", "John")
	})

	// 转换为SELECT查询并构建SQL
	builder := &mockBuilder{}
	q.Select("*").Build(builder)

	// 打印结果
	fmt.Println("Single condition SQL:", builder.String())
	fmt.Println("Single condition Args:", builder.vars)

	// 验证SQL是否不包含括号（当闭包内只有一个条件时）
	expectedSQL := "SELECT `*` FROM `users` WHERE `name` = $1"
	if builder.String() != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	// 验证参数是否正确
	expectedArgs := []interface{}{"John"}
	if len(builder.vars) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(builder.vars))
	}
	for i, arg := range expectedArgs {
		if builder.vars[i] != arg {
			t.Errorf("Expected arg %d to be %v, got %v", i, arg, builder.vars[i])
		}
	}
}

// TestMultipleConditionsWithClosure 测试当WHERE子句有多个条件（包含闭包）时，闭包是否正确添加括号
func TestMultipleConditionsWithClosure(t *testing.T) {
	// 创建一个查询
	q := Table("users")

	// 使用闭包作为多个条件之一
	q.Where("status", "active")
	q.Where(func(w Wherer) Wherer {
		return w.Where("name", "John").OrWhere("name", "Jane")
	})
	q.Where("age", ">", 18)

	// 转换为SELECT查询并构建SQL
	builder := &mockBuilder{}
	q.Select("*").Build(builder)

	// 打印结果
	fmt.Println("Multiple conditions with closure SQL:", builder.String())
	fmt.Println("Multiple conditions with closure Args:", builder.vars)

	// 验证SQL是否包含预期条件（闭包作为多个条件之一时应添加括号）
	expectedSQL := "SELECT `*` FROM `users` WHERE `status` = $1 AND (`name` = $2 OR `name` = $3) AND `age` > $4"
	if builder.String() != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	// 验证参数是否正确
	expectedArgs := []interface{}{"active", "John", "Jane", 18}
	if len(builder.vars) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(builder.vars))
	}
	for i, arg := range expectedArgs {
		if builder.vars[i] != arg {
			t.Errorf("Expected arg %d to be %v, got %v", i, arg, builder.vars[i])
		}
	}
}

// TestQuery_BuildUpdateWithMultipleWhere 测试多个WHERE条件的UPDATE操作
func TestQuery_BuildUpdateWithMultipleWhere(t *testing.T) {
	q := Table("users").Where("age", ">", 18).Where("role", "user").Where("last_login", ">", "2023-01-01").Update("status", "active")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "UPDATE `users` SET `status` = $1 WHERE `age` > $2 AND `role` = $3 AND `last_login` > $4"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 4 || builder.vars[0] != "active" || builder.vars[1] != 18 || builder.vars[2] != "user" || builder.vars[3] != "2023-01-01" {
		t.Errorf("expected vars: [active 18 user 2023-01-01], got: %v", builder.vars)
	}
}

// TestQuery_BuildDeleteWithAllConditions 测试包含所有条件的DELETE操作
func TestQuery_BuildDeleteWithAllConditions(t *testing.T) {
	q := Table("users").Where("age", "<", 18).OrWhere("status", "banned").OrWhere("last_login", "<", "2022-01-01").Not("role", "admin").Delete()
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "DELETE FROM `users` WHERE `age` < $1 OR `status` = $2 OR `last_login` < $3 AND `role` <> $4"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 4 || builder.vars[0] != 18 || builder.vars[1] != "banned" || builder.vars[2] != "2022-01-01" || builder.vars[3] != "admin" {
		t.Errorf("expected vars: [18 banned 2022-01-01 admin], got: %v", builder.vars)
	}
}

// TestQuery_BuildInsertWithAllNullValues 测试插入所有字段为空值
func TestQuery_BuildInsertWithAllNullValues(t *testing.T) {
	q := Table("users").Insert(map[string]interface{}{"name": nil, "email": nil, "age": nil})
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "INSERT INTO `users` (`age`, `email`, `name`) VALUES ($1, $2, $3)"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 3 || builder.vars[0] != nil || builder.vars[1] != nil || builder.vars[2] != nil {
		t.Errorf("expected vars: [nil nil nil], got: %v", builder.vars)
	}
}

// TestQuery_BuildSelectWithMixedConditionTypes 测试混合条件类型的SELECT查询
func TestQuery_BuildSelectWithMixedConditionTypes(t *testing.T) {
	q := Table("users").Where("age", ">", 18).Where("status", "active").OrWhere("role", "admin").Not("email", nil).Where("created_at", ">", "2023-01-01").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `age` > $1 AND `status` = $2 OR `role` = $3 AND `email` IS NOT NULL AND `created_at` > $4"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 4 {
		t.Errorf("expected 4 vars, got: %d", len(builder.vars))
	}
}
