package clause

import (
	"bytes"
	"fmt"
	"testing"
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

func (m *mockBuilder) AddVar(writer Writer, vars ...interface{}) {
	for i, v := range vars {
		m.vars = append(m.vars, v)
		if i > 0 {
			writer.WriteString(",")
		}
		writer.WriteString(fmt.Sprintf("$%d", len(m.vars)))
	}
}

func (m *mockBuilder) AddError(err error) error {
	if err != nil {
		m.errors = append(m.errors, err)
	}
	return err
}

// TestINExpression 测试IN表达式
func TestINExpression(t *testing.T) {
	tests := []struct {
		name         string
		in           IN
		expected     string
		expectedVars []interface{}
	}{
		{
			name:         "IN with multiple values",
			in:           IN{Column: "id", Values: []interface{}{1, 2, 3}},
			expected:     "`id` IN ($1,$2,$3)",
			expectedVars: []interface{}{1, 2, 3},
		},
		{
			name:         "IN with single value",
			in:           IN{Column: "id", Values: []interface{}{1}},
			expected:     "`id` = $1",
			expectedVars: []interface{}{1},
		},
		{
			name:         "IN with empty values",
			in:           IN{Column: "id", Values: []interface{}{}},
			expected:     "`id` IN (NULL)",
			expectedVars: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &mockBuilder{}
			tt.in.Build(builder)

			if builder.String() != tt.expected {
				t.Errorf("expected SQL: %s, got: %s", tt.expected, builder.String())
			}

			if len(builder.vars) != len(tt.expectedVars) {
				t.Errorf("expected %d vars, got %d", len(tt.expectedVars), len(builder.vars))
			} else {
				for i := range builder.vars {
					if builder.vars[i] != tt.expectedVars[i] {
						t.Errorf("var at index %d: expected %v, got %v", i, tt.expectedVars[i], builder.vars[i])
					}
				}
			}
		})
	}
}

// TestINNegationExpression 测试IN表达式的NegationBuild方法
func TestINNegationExpression(t *testing.T) {
	tests := []struct {
		name         string
		in           IN
		expected     string
		expectedVars []interface{}
	}{
		{
			name:         "NOT IN with multiple values",
			in:           IN{Column: "id", Values: []interface{}{1, 2, 3}},
			expected:     "`id` NOT IN ($1,$2,$3)",
			expectedVars: []interface{}{1, 2, 3},
		},
		{
			name:         "NOT IN with single value",
			in:           IN{Column: "id", Values: []interface{}{1}},
			expected:     "`id` <> $1",
			expectedVars: []interface{}{1},
		},
		{
			name:         "NOT IN with empty values",
			in:           IN{Column: "id", Values: []interface{}{}},
			expected:     "`id` IS NOT NULL",
			expectedVars: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &mockBuilder{}
			tt.in.NegationBuild(builder)

			if builder.String() != tt.expected {
				t.Errorf("expected SQL: %s, got: %s", tt.expected, builder.String())
			}

			if len(builder.vars) != len(tt.expectedVars) {
				t.Errorf("expected %d vars, got %d", len(tt.expectedVars), len(builder.vars))
			} else {
				for i := range builder.vars {
					if builder.vars[i] != tt.expectedVars[i] {
						t.Errorf("var at index %d: expected %v, got %v", i, tt.expectedVars[i], builder.vars[i])
					}
				}
			}
		})
	}
}

// TestEqExpression 测试Eq表达式
func TestEqExpression(t *testing.T) {
	tests := []struct {
		name         string
		eq           Eq
		expected     string
		expectedVars []interface{}
	}{
		{
			name:         "Eq with string value",
			eq:           Eq{Column: "name", Value: "test"},
			expected:     "`name` = $1",
			expectedVars: []interface{}{"test"},
		},
		{
			name:         "Eq with numeric value",
			eq:           Eq{Column: "age", Value: 18},
			expected:     "`age` = $1",
			expectedVars: []interface{}{18},
		},
		{
			name:         "Eq with nil value",
			eq:           Eq{Column: "email", Value: nil},
			expected:     "`email` IS NULL",
			expectedVars: []interface{}{},
		},
		{
			name:         "Eq with slice value",
			eq:           Eq{Column: "id", Value: []interface{}{1, 2, 3}},
			expected:     "`id` IN ($1,$2,$3)",
			expectedVars: []interface{}{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &mockBuilder{}
			tt.eq.Build(builder)

			if builder.String() != tt.expected {
				t.Errorf("expected SQL: %s, got: %s", tt.expected, builder.String())
			}

			if len(builder.vars) != len(tt.expectedVars) {
				t.Errorf("expected %d vars, got %d", len(tt.expectedVars), len(builder.vars))
			} else {
				for i := range builder.vars {
					if builder.vars[i] != tt.expectedVars[i] {
						t.Errorf("var at index %d: expected %v, got %v", i, tt.expectedVars[i], builder.vars[i])
					}
				}
			}
		})
	}
}

// TestNeqExpression 测试Neq表达式
func TestNeqExpression(t *testing.T) {
	tests := []struct {
		name         string
		neq          Neq
		expected     string
		expectedVars []interface{}
	}{
		{
			name:         "Neq with string value",
			neq:          Neq{Column: "name", Value: "test"},
			expected:     "`name` <> $1",
			expectedVars: []interface{}{"test"},
		},
		{
			name:         "Neq with numeric value",
			neq:          Neq{Column: "age", Value: 18},
			expected:     "`age` <> $1",
			expectedVars: []interface{}{18},
		},
		{
			name:         "Neq with nil value",
			neq:          Neq{Column: "email", Value: nil},
			expected:     "`email` IS NOT NULL",
			expectedVars: []interface{}{},
		},
		{
			name:         "Neq with slice value",
			neq:          Neq{Column: "id", Value: []interface{}{1, 2, 3}},
			expected:     "`id` NOT IN ($1,$2,$3)",
			expectedVars: []interface{}{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &mockBuilder{}
			tt.neq.Build(builder)

			if builder.String() != tt.expected {
				t.Errorf("expected SQL: %s, got: %s", tt.expected, builder.String())
			}

			if len(builder.vars) != len(tt.expectedVars) {
				t.Errorf("expected %d vars, got %d", len(tt.expectedVars), len(builder.vars))
			} else {
				for i := range builder.vars {
					if builder.vars[i] != tt.expectedVars[i] {
						t.Errorf("var at index %d: expected %v, got %v", i, tt.expectedVars[i], builder.vars[i])
					}
				}
			}
		})
	}
}

// TestComparisonExpressions 测试比较表达式
func TestComparisonExpressions(t *testing.T) {
	tests := []struct {
		name         string
		expr         Expression
		expected     string
		expectedVars []interface{}
	}{
		{
			name:         "Gt expression",
			expr:         Gt{Column: "age", Value: 18},
			expected:     "`age` > $1",
			expectedVars: []interface{}{18},
		},
		{
			name:         "Gte expression",
			expr:         Gte{Column: "age", Value: 18},
			expected:     "`age` >= $1",
			expectedVars: []interface{}{18},
		},
		{
			name:         "Lt expression",
			expr:         Lt{Column: "age", Value: 30},
			expected:     "`age` < $1",
			expectedVars: []interface{}{30},
		},
		{
			name:         "Lte expression",
			expr:         Lte{Column: "age", Value: 30},
			expected:     "`age` <= $1",
			expectedVars: []interface{}{30},
		},
		{
			name:         "Like expression",
			expr:         Like{Column: "name", Value: "%test%"},
			expected:     "`name` LIKE $1",
			expectedVars: []interface{}{"%test%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &mockBuilder{}
			tt.expr.Build(builder)

			if builder.String() != tt.expected {
				t.Errorf("expected SQL: %s, got: %s", tt.expected, builder.String())
			}

			if len(builder.vars) != len(tt.expectedVars) {
				t.Errorf("expected %d vars, got %d", len(tt.expectedVars), len(builder.vars))
			} else {
				for i := range builder.vars {
					if builder.vars[i] != tt.expectedVars[i] {
						t.Errorf("var at index %d: expected %v, got %v", i, tt.expectedVars[i], builder.vars[i])
					}
				}
			}
		})
	}
}

// TestAndExpression 测试And表达式
func TestAndExpression(t *testing.T) {
	tests := []struct {
		name         string
		expr         Expression
		expected     string
		expectedVars []interface{}
	}{
		{
			name:         "And with multiple conditions",
			expr:         And(Eq{Column: "name", Value: "test"}, Gt{Column: "age", Value: 18}),
			expected:     "(`name` = $1 AND `age` > $2)",
			expectedVars: []interface{}{"test", 18},
		},
		{
			name:         "And with single condition",
			expr:         And(Eq{Column: "name", Value: "test"}),
			expected:     "`name` = $1",
			expectedVars: []interface{}{"test"},
		},
		{
			name:         "And with nested Or",
			expr:         And(Eq{Column: "status", Value: "active"}, Or(Eq{Column: "age", Value: 18}, Eq{Column: "age", Value: 19})),
			expected:     "(`status` = $1 AND (`age` = $2 OR `age` = $3))",
			expectedVars: []interface{}{"active", 18, 19},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &mockBuilder{}
			tt.expr.Build(builder)

			if builder.String() != tt.expected {
				t.Errorf("expected SQL: %s, got: %s", tt.expected, builder.String())
			}

			if len(builder.vars) != len(tt.expectedVars) {
				t.Errorf("expected %d vars, got %d", len(tt.expectedVars), len(builder.vars))
			} else {
				for i := range builder.vars {
					if builder.vars[i] != tt.expectedVars[i] {
						t.Errorf("var at index %d: expected %v, got %v", i, tt.expectedVars[i], builder.vars[i])
					}
				}
			}
		})
	}
}

// TestOrExpression 测试Or表达式
func TestOrExpression(t *testing.T) {
	tests := []struct {
		name         string
		expr         Expression
		expected     string
		expectedVars []interface{}
	}{
		{
			name:         "Or with multiple conditions",
			expr:         Or(Eq{Column: "name", Value: "test"}, Eq{Column: "name", Value: "admin"}),
			expected:     "(`name` = $1 OR `name` = $2)",
			expectedVars: []interface{}{"test", "admin"},
		},
		{
			name:         "Or with single condition",
			expr:         Or(Eq{Column: "name", Value: "test"}),
			expected:     "`name` = $1",
			expectedVars: []interface{}{"test"},
		},
		{
			name:         "Or with nested And",
			expr:         Or(And(Eq{Column: "name", Value: "test"}, Gt{Column: "age", Value: 18}), And(Eq{Column: "name", Value: "admin"}, Gt{Column: "age", Value: 25})),
			expected:     "((`name` = $1 AND `age` > $2) OR (`name` = $3 AND `age` > $4))",
			expectedVars: []interface{}{"test", 18, "admin", 25},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &mockBuilder{}
			tt.expr.Build(builder)

			if builder.String() != tt.expected {
				t.Errorf("expected SQL: %s, got: %s", tt.expected, builder.String())
			}

			if len(builder.vars) != len(tt.expectedVars) {
				t.Errorf("expected %d vars, got %d", len(tt.expectedVars), len(builder.vars))
			} else {
				for i := range builder.vars {
					if builder.vars[i] != tt.expectedVars[i] {
						t.Errorf("var at index %d: expected %v, got %v", i, tt.expectedVars[i], builder.vars[i])
					}
				}
			}
		})
	}
}

// TestNotExpression 测试Not表达式
func TestNotExpression(t *testing.T) {
	tests := []struct {
		name         string
		expr         Expression
		expected     string
		expectedVars []interface{}
	}{
		{
			name:         "Not with Eq",
			expr:         Not(Eq{Column: "name", Value: "test"}),
			expected:     "`name` <> $1",
			expectedVars: []interface{}{"test"},
		},
		{
			name:         "Not with IN",
			expr:         Not(IN{Column: "id", Values: []interface{}{1, 2, 3}}),
			expected:     "`id` NOT IN ($1,$2,$3)",
			expectedVars: []interface{}{1, 2, 3},
		},
		{
			name:         "Not with And",
			expr:         Not(And(Eq{Column: "name", Value: "test"}, Gt{Column: "age", Value: 18})),
			expected:     "(`name` <> $1 AND `age` <= $2)",
			expectedVars: []interface{}{"test", 18},
		},
		{
			name:         "Not with Or",
			expr:         Not(Or(Eq{Column: "name", Value: "test"}, Eq{Column: "name", Value: "admin"})),
			expected:     "NOT (`name` = $1 OR `name` = $2)",
			expectedVars: []interface{}{"test", "admin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &mockBuilder{}
			tt.expr.Build(builder)

			if builder.String() != tt.expected {
				t.Errorf("expected SQL: %s, got: %s", tt.expected, builder.String())
			}

			if len(builder.vars) != len(tt.expectedVars) {
				t.Errorf("expected %d vars, got %d", len(tt.expectedVars), len(builder.vars))
			} else {
				for i := range builder.vars {
					if builder.vars[i] != tt.expectedVars[i] {
						t.Errorf("var at index %d: expected %v, got %v", i, tt.expectedVars[i], builder.vars[i])
					}
				}
			}
		})
	}
}
