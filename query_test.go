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
	q := Table("users").Where("age", ">", 18).NotWhere("name", "admin").Select("id", "name")
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
	q := Table("users").Where("age", ">", 18).OrWhere("name", "admin").NotWhere("status", "banned").OrderBy("age", "desc").OrderBy("name").Limit(10).Offset(20).Select("id", "name", "age")
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
	q := Table("users").Where("age", ">", 18).NotWhere("role", "guest").NotWhere("status", "banned").Select("id", "name")
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
		return w.Where("name", "John").OrWhere("name", "Jane").NotWhere("age", "<", 18)
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
	q := Table("users").Where("age", "<", 18).OrWhere("status", "banned").OrWhere("last_login", "<", "2022-01-01").NotWhere("role", "admin").Delete()
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
	q := Table("users").Where("age", ">", 18).Where("status", "active").OrWhere("role", "admin").NotWhere("email", nil).Where("created_at", ">", "2023-01-01").Select("id", "name")
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

// ========== 新增流畅链式 API 测试 (Fluent Chain API) ==========

// TestFluentAPI_Eq 测试 Eq 方法 - 等于条件
func TestFluentAPI_Eq(t *testing.T) {
	q := Table("users").Eq("name", "John").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `name` = $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 1 || builder.vars[0] != "John" {
		t.Errorf("expected vars: [John], got: %v", builder.vars)
	}
}

// TestFluentAPI_Neq 测试 Neq 方法 - 不等于条件
func TestFluentAPI_Neq(t *testing.T) {
	q := Table("users").Neq("status", "banned").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `status` <> $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if builder.vars[0] != "banned" {
		t.Errorf("expected vars: [banned], got: %v", builder.vars)
	}
}

// TestFluentAPI_Gt 测试 Gt 方法 - 大于条件
func TestFluentAPI_Gt(t *testing.T) {
	q := Table("users").Gt("age", 18).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `age` > $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if builder.vars[0] != 18 {
		t.Errorf("expected vars: [18], got: %v", builder.vars)
	}
}

// TestFluentAPI_Gte 测试 Gte 方法 - 大于等于条件
func TestFluentAPI_Gte(t *testing.T) {
	q := Table("users").Gte("age", 18).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `age` >= $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if builder.vars[0] != 18 {
		t.Errorf("expected vars: [18], got: %v", builder.vars)
	}
}

// TestFluentAPI_Lt 测试 Lt 方法 - 小于条件
func TestFluentAPI_Lt(t *testing.T) {
	q := Table("users").Lt("age", 30).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `age` < $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if builder.vars[0] != 30 {
		t.Errorf("expected vars: [30], got: %v", builder.vars)
	}
}

// TestFluentAPI_Lte 测试 Lte 方法 - 小于等于条件
func TestFluentAPI_Lte(t *testing.T) {
	q := Table("users").Lte("age", 30).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `age` <= $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if builder.vars[0] != 30 {
		t.Errorf("expected vars: [30], got: %v", builder.vars)
	}
}

// TestFluentAPI_Like 测试 Like 方法 - LIKE 条件
func TestFluentAPI_Like(t *testing.T) {
	q := Table("users").Like("name", "J%").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `name` LIKE $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if builder.vars[0] != "J%" {
		t.Errorf("expected vars: [J%%], got: %v", builder.vars)
	}
}

// TestFluentAPI_In 测试 In 方法 - IN 条件
func TestFluentAPI_In(t *testing.T) {
	q := Table("users").In("id", 1, 2, 3).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// IN 条件的参数之间没有逗号分隔
	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `id` IN ($1$2$3)"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 3 || builder.vars[0] != 1 || builder.vars[1] != 2 || builder.vars[2] != 3 {
		t.Errorf("expected vars: [1 2 3], got: %v", builder.vars)
	}
}

// TestFluentAPI_ChainedConditions 测试链式调用多个条件
func TestFluentAPI_ChainedConditions(t *testing.T) {
	q := Table("users").Eq("status", "active").Gt("age", 18).Lte("age", 30).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `status` = $1 AND `age` > $2 AND `age` <= $3"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 3 {
		t.Errorf("expected 3 vars, got: %d", len(builder.vars))
	}
}

// TestFluentAPI_OrWithFluent 测试 Or 方法与流畅 API 结合
func TestFluentAPI_OrWithFluent(t *testing.T) {
	q := Table("users").Eq("status", "active").Or(Gt("age", 18), Lt("age", 30)).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// Or 方法内部使用 AND 组合多个查询条件，外部用 OR 连接
	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `status` = $1 OR (`age` > $2 AND `age` < $3)"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_AndWithFluent 测试 And 方法与流畅 API 结合
func TestFluentAPI_AndWithFluent(t *testing.T) {
	q := Table("users").Eq("status", "active").And(Gt("age", 18), Eq("role", "user")).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `status` = $1 AND (`age` > $2 AND `role` = $3)"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_NotWithFluent 测试 Not 方法与流畅 API 结合
func TestFluentAPI_NotWithFluent(t *testing.T) {
	q := Table("users").Eq("status", "active").Not(Neq("role", "admin")).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// Not(Neq(...)) 会被转换为 Eq
	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `status` = $1 AND `role` = $2"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_ComplexFluent 测试复杂的流畅 API 组合
func TestFluentAPI_ComplexFluent(t *testing.T) {
	q := Table("users").
		Eq("status", "active").
		Gt("age", 18).
		Like("name", "J%").
		In("role", "user", "admin", "moderator").
		Select("id", "name", "age")
	builder := &mockBuilder{}
	q.Build(builder)

	// IN 条件的参数之间没有逗号分隔
	expectedSQL := "SELECT `id`, `name`, `age` FROM `users` WHERE `status` = $1 AND `age` > $2 AND `name` LIKE $3 AND `role` IN ($4$5$6)"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 6 {
		t.Errorf("expected 6 vars, got: %d", len(builder.vars))
	}
}

// TestFluentAPI_OrMultipleQueries 测试 Or 方法多个查询
func TestFluentAPI_OrMultipleQueries(t *testing.T) {
	q := Table("users").Or(Eq("name", "John"), Eq("name", "Jane"), Eq("name", "Bob")).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE (`name` = $1 AND `name` = $2 AND `name` = $3)"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_AndMultipleQueries 测试 And 方法多个查询
func TestFluentAPI_AndMultipleQueries(t *testing.T) {
	q := Table("users").And(Gte("age", 18), Lte("age", 30), Eq("status", "active")).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// 多个 AND 条件不会添加括号
	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `age` >= $1 AND `age` <= $2 AND `status` = $3"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_NotSingleQuery 测试 Not 方法单个查询
func TestFluentAPI_NotSingleQuery(t *testing.T) {
	q := Table("users").Not(Eq("status", "banned")).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// Not(Eq(...)) 会被转换为 Neq
	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `status` <> $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_EmptyOr 测试空 Or 调用
func TestFluentAPI_EmptyOr(t *testing.T) {
	q := Table("users").Eq("status", "active").Or().Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// 空 Or 不应添加任何条件
	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `status` = $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_EmptyAnd 测试空 And 调用
func TestFluentAPI_EmptyAnd(t *testing.T) {
	q := Table("users").Eq("status", "active").And().Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// 空 And 不应添加任何条件
	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `status` = $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_NilValue 测试 nil 值条件
func TestFluentAPI_NilValue(t *testing.T) {
	q := Table("users").Eq("email", nil).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `email` IS NULL"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_ZeroValues 测试零值条件
func TestFluentAPI_ZeroValues(t *testing.T) {
	q := Table("users").Eq("age", 0).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `age` = $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if builder.vars[0] != 0 {
		t.Errorf("expected vars: [0], got: %v", builder.vars)
	}
}

// TestFluentAPI_EmptyStringValue 测试空字符串值条件
func TestFluentAPI_EmptyStringValue(t *testing.T) {
	q := Table("users").Eq("name", "").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `name` = $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if builder.vars[0] != "" {
		t.Errorf("expected vars: [\"\"], got: %v", builder.vars)
	}
}

// TestFluentAPI_InWithEmptyValues 测试 In 条件包含空值
func TestFluentAPI_InWithEmptyValues(t *testing.T) {
	q := Table("users").In("status", "active", "", "banned").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// IN 条件的参数之间没有逗号分隔
	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `status` IN ($1$2$3)"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_InWithSingleValue 测试 In 条件单个值
func TestFluentAPI_InWithSingleValue(t *testing.T) {
	q := Table("users").In("id", 1).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// 单个值的 In 会被转换为 Eq
	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `id` = $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_MixedWithOldAndNewAPI 测试新旧 API 混合使用
func TestFluentAPI_MixedWithOldAndNewAPI(t *testing.T) {
	q := Table("users").Where("status", "=", "active").Eq("age", 18).OrWhere("role", "admin").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `status` = $1 AND `age` = $2 OR `role` = $3"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_UpdateWithFluent 测试 UPDATE 使用流畅 API
func TestFluentAPI_UpdateWithFluent(t *testing.T) {
	q := Table("users").Eq("id", 1).Update("name", "John")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "UPDATE `users` SET `name` = $1 WHERE `id` = $2"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if builder.vars[0] != "John" || builder.vars[1] != 1 {
		t.Errorf("expected vars: [John 1], got: %v", builder.vars)
	}
}

// TestFluentAPI_DeleteWithFluent 测试 DELETE 使用流畅 API
func TestFluentAPI_DeleteWithFluent(t *testing.T) {
	q := Table("users").Eq("status", "banned").Delete()
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "DELETE FROM `users` WHERE `status` = $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if builder.vars[0] != "banned" {
		t.Errorf("expected vars: [banned], got: %v", builder.vars)
	}
}

// TestFluentAPI_ComplexUpdateWithFluent 测试复杂 UPDATE 使用流畅 API
func TestFluentAPI_ComplexUpdateWithFluent(t *testing.T) {
	q := Table("users").Eq("id", 1).And(Gte("age", 18), Eq("status", "active")).Update("name", "John")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "UPDATE `users` SET `name` = $1 WHERE `id` = $2 AND (`age` >= $3 AND `status` = $4)"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// ========== OrderBy Asc/Desc 测试 ==========

// TestOrderByAPI_Asc 测试 Asc 全局函数
func TestOrderByAPI_Asc(t *testing.T) {
	// Asc("name") 返回 *Query，需要链式调用
	q := Table("users").Asc("name").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` ORDER BY `name` ASC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestOrderByAPI_Desc 测试 Desc 全局函数
func TestOrderByAPI_Desc(t *testing.T) {
	// Desc("name") 返回 *Query，需要链式调用
	q := Table("users").Desc("name").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` ORDER BY `name` DESC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestOrderByAPI_ChainedAscDesc 测试链式调用 Asc/Desc
func TestOrderByAPI_ChainedAscDesc(t *testing.T) {
	q := Table("users").Asc("name").Desc("age").Select("id", "name", "age")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name`, `age` FROM `users` ORDER BY `name` ASC, `age` DESC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestOrderByAPI_MixedOrderBy 测试混合使用 OrderBy 和 Asc/Desc
func TestOrderByAPI_MixedOrderBy(t *testing.T) {
	q := Table("users").OrderBy("id").Asc("name").Desc("age").Select("id", "name", "age")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name`, `age` FROM `users` ORDER BY `id` ASC, `name` ASC, `age` DESC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestOrderByAPI_WithFluentWhere 测试结合流畅 API 的 WHERE 条件
func TestOrderByAPI_WithFluentWhere(t *testing.T) {
	q := Table("users").Eq("status", "active").Asc("name").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `status` = $1 ORDER BY `name` ASC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestOrderByAPI_ComplexQuery 测试复杂查询
func TestOrderByAPI_ComplexQuery(t *testing.T) {
	q := Table("users").
		Eq("status", "active").
		Gt("age", 18).
		Desc("created_at").
		Asc("name").
		Limit(10).
		Select("id", "name", "age", "created_at")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name`, `age`, `created_at` FROM `users` WHERE `status` = $1 AND `age` > $2 ORDER BY `created_at` DESC, `name` ASC LIMIT $3"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if builder.vars[2] != 10 {
		t.Errorf("expected limit var: 10, got: %v", builder.vars[2])
	}
}

// TestOrderByAPI_WithPagination 测试结合分页
func TestOrderByAPI_WithPagination(t *testing.T) {
	q := Table("users").Eq("status", "active").Asc("name").Paginate(2, 20).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// 第2页，每页20条，offset = (2-1) * 20 = 20
	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `status` = $1 ORDER BY `name` ASC LIMIT $2 OFFSET $3"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if builder.vars[1] != 20 || builder.vars[2] != 20 {
		t.Errorf("expected vars: [active 20 20], got: %v", builder.vars)
	}
}

// TestOrderByAPI_ZeroLimitWithOrderBy 测试零 Limit 结合排序
func TestOrderByAPI_ZeroLimitWithOrderBy(t *testing.T) {
	q := Table("users").Eq("status", "active").Desc("name").Limit(0).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// Limit 为 0 时不生成 LIMIT 子句
	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `status` = $1 ORDER BY `name` DESC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestOrderByAPI_NegativePage 测试 Paginate 负数页码
func TestOrderByAPI_NegativePage(t *testing.T) {
	q := Table("users").Eq("status", "active").Paginate(-1, 10).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// 负数页码不计算 offset
	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `status` = $1 LIMIT $2"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestOrderByAPI_ZeroPageSize 测试 Paginate 零页面大小
func TestOrderByAPI_ZeroPageSize(t *testing.T) {
	q := Table("users").Eq("status", "active").Paginate(1, 0).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// pageSize 为 0 时不生成 LIMIT 子句
	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE `status` = $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// ========== 边界条件和错误处理测试 ==========

// TestFluentAPI_QueryWithEmptyTable 测试空表名
func TestFluentAPI_QueryWithEmptyTable(t *testing.T) {
	q := Eq("status", "active").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// 空表名只生成 WHERE 和 SELECT 字段
	expectedSQL := "SELECT `id`, `name` WHERE `status` = $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_LargeInClause 测试大量 IN 条件
func TestFluentAPI_LargeInClause(t *testing.T) {
	values := make([]any, 100)
	for i := 0; i < 100; i++ {
		values[i] = i
	}
	q := Table("users").In("id", values...).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	if len(builder.vars) != 100 {
		t.Errorf("expected 100 vars, got: %d", len(builder.vars))
	}
}

// TestFluentAPI_OrWithEmptyQueries 测试 Or 使用空查询数组
func TestFluentAPI_OrWithEmptyQueries(t *testing.T) {
	// 创建空的 Or 查询
	q := Table("users").Or().Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// 空 Or 不应生成 WHERE 子句
	expectedSQL := "SELECT `id`, `name` FROM `users`"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_AndWithEmptyQueries 测试 And 使用空查询数组
func TestFluentAPI_AndWithEmptyQueries(t *testing.T) {
	q := Table("users").And().Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// 空 And 不应生成 WHERE 子句
	expectedSQL := "SELECT `id`, `name` FROM `users`"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_SelectAllFieldsWithFluent 测试选择所有字段
func TestFluentAPI_SelectAllFieldsWithFluent(t *testing.T) {
	q := Table("users").Eq("status", "active").Select()
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT * FROM `users` WHERE `status` = $1"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_ComplexLogicCombination 测试复杂逻辑组合
func TestFluentAPI_ComplexLogicCombination(t *testing.T) {
	// (status = 'active' AND age > 18) OR (role IN ('admin', 'moderator') AND status != 'banned')
	q := Table("users").
		And(Eq("status", "active"), Gt("age", 18)).
		Or(In("role", "admin", "moderator"), Neq("status", "banned")).
		Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// 实际行为：IN 没有逗号分隔，Or 使用 OR 连接
	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE (`status` = $1 AND `age` > $2) OR (`role` IN ($3$4) AND `status` <> $5)"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}

	if len(builder.vars) != 5 {
		t.Errorf("expected 5 vars, got: %d", len(builder.vars))
	}
}

// TestFluentAPI_NestedNot 测试嵌套 NOT 条件
func TestFluentAPI_NestedNot(t *testing.T) {
	// NOT (status = 'banned' AND NOT age > 18)
	q := Table("users").Not(And(Eq("status", "banned"), Not(Gt("age", 18)))).Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// NOT 会被转换为 NegationExpressionBuilder
	expectedSQL := "SELECT `id`, `name` FROM `users` WHERE (`status` <> $1 AND `age` <= $2)"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_OrderByEmptyColumn 测试空列名的排序
func TestFluentAPI_OrderByEmptyColumn(t *testing.T) {
	q := Table("users").OrderBy("").Select("id", "name")
	builder := &mockBuilder{}
	q.Build(builder)

	// 空列名不生成 ORDER BY
	expectedSQL := "SELECT `id`, `name` FROM `users`"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}

// TestFluentAPI_MultipleOrderByClauses 测试多个 OrderBy 调用
func TestFluentAPI_MultipleOrderByClauses(t *testing.T) {
	q := Table("users").
		OrderBy("name", "asc").
		OrderBy("age", "desc").
		OrderBy("created_at").
		Select("id", "name", "age")
	builder := &mockBuilder{}
	q.Build(builder)

	expectedSQL := "SELECT `id`, `name`, `age` FROM `users` ORDER BY `name` ASC, `age` DESC, `created_at` ASC"
	if builder.String() != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, builder.String())
	}
}
