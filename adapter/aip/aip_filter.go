package aip

import (
	"fmt"

	"github.com/epkgs/query"
	"github.com/epkgs/query/clause"

	filtering "go.einride.tech/aip/filtering"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// FromFilter 将 AIP Filter 转换为 clause.Where
func FromFilter(filter filtering.Filter) (clause.Where, error) {
	// 创建 Query 对象
	wherer := query.NewWhereBuilder()

	// 解析表达式，处理空Filter情况
	checkedExpr := filter.CheckedExpr
	if checkedExpr != nil {
		exp := checkedExpr.GetExpr()
		if exp != nil {
			// 解析表达式
			exprs, err := parseExpr(exp)
			if err != nil {
				return clause.Where{}, err
			}

			isOr := false
			if len(exprs) == 1 {
				if andExpr, ok := exprs[0].(clause.AndExpr); ok {
					exprs = andExpr.Exprs
				} else if orExpr, ok := exprs[0].(clause.OrExpr); ok {
					exprs = orExpr.Exprs
					isOr = true
				}
			}

			for _, e := range exprs {
				if isOr {
					wherer.OrWhere(e)
				} else {
					wherer.Where(e)
				}
				if wherer.Error != nil {
					return clause.Where{}, wherer.Error
				}
			}
		}
	}

	return wherer.WhereExpr(), nil
}

// parseExpr 将 *exprpb.Expr 转换为 []clause.Expression
func parseExpr(exp *exprpb.Expr) ([]clause.Expression, error) {
	if exp == nil {
		return nil, fmt.Errorf("nil expression")
	}

	switch kind := exp.ExprKind.(type) {
	case *exprpb.Expr_CallExpr:
		return parseCallExpr(kind.CallExpr)
	case *exprpb.Expr_IdentExpr:
		// 处理标识符表达式（字段名）
		fieldName := kind.IdentExpr.Name
		return []clause.Expression{clause.Eq{Column: fieldName, Value: true}}, nil
	default:
		// TODO: 支持更多类型
		return nil, fmt.Errorf("unsupported expression type: %T", kind)
	}
}

// parseCallExpr 解析函数调用表达式
func parseCallExpr(call *exprpb.Expr_Call) ([]clause.Expression, error) {
	if call == nil {
		return nil, fmt.Errorf("nil call expression")
	}

	// 获取函数名
	funcName := call.Function

	// 处理运算符映射
	// 在 AIP filtering 中，运算符被建模为函数调用
	switch funcName {
	case "AND":
		return parseAndExpr(call.Args)
	case "OR":
		return parseOrExpr(call.Args)
	case "EQUALS", "=":
		return parseEqualsExpr(call.Args)
	case "NOT_EQUALS", "!=":
		return parseNotEqualsExpr(call.Args)
	case "GREATER_THAN", ">":
		return parseGreaterThanExpr(call.Args)
	case "GREATER_EQUALS", ">=":
		return parseGreaterEqualsExpr(call.Args)
	case "LESS_THAN", "<":
		return parseLessThanExpr(call.Args)
	case "LESS_EQUALS", "<=":
		return parseLessEqualsExpr(call.Args)
	case "NOT":
		return parseNotExpr(call.Args)
	case "HAS", "IN":
		return parseHasExpr(call.Args)
	default:
		return nil, fmt.Errorf("unsupported function: %s", funcName)
	}
}

// parseAndExpr 解析 AND 函数
func parseAndExpr(args []*exprpb.Expr) ([]clause.Expression, error) {
	if len(args) == 0 {
		return nil, nil
	}

	var exprs []clause.Expression
	for _, arg := range args {
		parsed, err := parseExpr(arg)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, parsed...)
	}

	if len(exprs) == 1 {
		return exprs, nil
	}

	return []clause.Expression{clause.And(exprs...)}, nil
}

// parseOrExpr 解析 OR 函数
func parseOrExpr(args []*exprpb.Expr) ([]clause.Expression, error) {
	if len(args) == 0 {
		return nil, nil
	}

	var exprs []clause.Expression
	for _, arg := range args {
		parsed, err := parseExpr(arg)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, parsed...)
	}

	if len(exprs) == 1 {
		return exprs, nil
	}

	return []clause.Expression{clause.Or(exprs...)}, nil
}

// parseField 解析字段表达式
func parseField(exp *exprpb.Expr) (string, error) {
	if exp == nil {
		return "", fmt.Errorf("nil expression")
	}

	// 检查是否是标识符表达式
	if ident := exp.GetIdentExpr(); ident != nil {
		return ident.Name, nil
	}

	// 检查是否是选择表达式
	if selectExpr := exp.GetSelectExpr(); selectExpr != nil {
		return selectExpr.Field, nil
	}

	return "", fmt.Errorf("unsupported field expression: %T", exp.ExprKind)
}

// parseValue 解析值表达式
func parseValue(exp *exprpb.Expr) (interface{}, error) {
	if exp == nil {
		return nil, fmt.Errorf("nil expression")
	}

	// 检查是否是常量表达式
	if constExpr := exp.GetConstExpr(); constExpr != nil {
		// 根据常量类型返回不同的值
		switch kind := constExpr.ConstantKind.(type) {
		case *exprpb.Constant_StringValue:
			return kind.StringValue, nil
		case *exprpb.Constant_Int64Value:
			return kind.Int64Value, nil
		case *exprpb.Constant_Uint64Value:
			return kind.Uint64Value, nil
		case *exprpb.Constant_DoubleValue:
			return kind.DoubleValue, nil
		case *exprpb.Constant_BoolValue:
			return kind.BoolValue, nil
		case *exprpb.Constant_NullValue:
			return nil, nil
		default:
			return nil, fmt.Errorf("unsupported constant type: %T", kind)
		}
	}

	return nil, fmt.Errorf("unsupported value expression: %T", exp.ExprKind)
}

// parseEqualsExpr 解析 EQUALS 函数
func parseEqualsExpr(args []*exprpb.Expr) ([]clause.Expression, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("EQUALS expects exactly 2 arguments, got %d", len(args))
	}

	// 解析字段名
	field, err := parseField(args[0])
	if err != nil {
		return nil, err
	}

	// 解析值
	value, err := parseValue(args[1])
	if err != nil {
		return nil, err
	}

	return []clause.Expression{clause.Eq{Column: field, Value: value}}, nil
}

// parseNotEqualsExpr 解析 NOT_EQUALS 函数
func parseNotEqualsExpr(args []*exprpb.Expr) ([]clause.Expression, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("NOT_EQUALS expects exactly 2 arguments, got %d", len(args))
	}

	// 解析字段名
	field, err := parseField(args[0])
	if err != nil {
		return nil, err
	}

	// 解析值
	value, err := parseValue(args[1])
	if err != nil {
		return nil, err
	}

	return []clause.Expression{clause.Neq{Column: field, Value: value}}, nil
}

// parseGreaterThanExpr 解析 GREATER_THAN 函数
func parseGreaterThanExpr(args []*exprpb.Expr) ([]clause.Expression, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("GREATER_THAN expects exactly 2 arguments, got %d", len(args))
	}

	// 解析字段名
	field, err := parseField(args[0])
	if err != nil {
		return nil, err
	}

	// 解析值
	value, err := parseValue(args[1])
	if err != nil {
		return nil, err
	}

	return []clause.Expression{clause.Gt{Column: field, Value: value}}, nil
}

// parseGreaterEqualsExpr 解析 GREATER_EQUALS 函数
func parseGreaterEqualsExpr(args []*exprpb.Expr) ([]clause.Expression, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("GREATER_EQUALS expects exactly 2 arguments, got %d", len(args))
	}

	// 解析字段名
	field, err := parseField(args[0])
	if err != nil {
		return nil, err
	}

	// 解析值
	value, err := parseValue(args[1])
	if err != nil {
		return nil, err
	}

	return []clause.Expression{clause.Gte{Column: field, Value: value}}, nil
}

// parseLessThanExpr 解析 LESS_THAN 函数
func parseLessThanExpr(args []*exprpb.Expr) ([]clause.Expression, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("LESS_THAN expects exactly 2 arguments, got %d", len(args))
	}

	// 解析字段名
	field, err := parseField(args[0])
	if err != nil {
		return nil, err
	}

	// 解析值
	value, err := parseValue(args[1])
	if err != nil {
		return nil, err
	}

	return []clause.Expression{clause.Lt{Column: field, Value: value}}, nil
}

// parseLessEqualsExpr 解析 LESS_EQUALS 函数
func parseLessEqualsExpr(args []*exprpb.Expr) ([]clause.Expression, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("LESS_EQUALS expects exactly 2 arguments, got %d", len(args))
	}

	// 解析字段名
	field, err := parseField(args[0])
	if err != nil {
		return nil, err
	}

	// 解析值
	value, err := parseValue(args[1])
	if err != nil {
		return nil, err
	}

	return []clause.Expression{clause.Lte{Column: field, Value: value}}, nil
}

// parseNotExpr 解析 NOT 函数
func parseNotExpr(args []*exprpb.Expr) ([]clause.Expression, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("NOT expects exactly 1 argument, got %d", len(args))
	}

	parsed, err := parseExpr(args[0])
	if err != nil {
		return nil, err
	}

	if len(parsed) != 1 {
		return nil, fmt.Errorf("NOT operand must be a single expression")
	}

	return []clause.Expression{clause.Not(parsed...)}, nil
}

// parseHasExpr 解析 HAS 函数
func parseHasExpr(args []*exprpb.Expr) ([]clause.Expression, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("HAS expects exactly 2 arguments, got %d", len(args))
	}

	// 解析字段名
	field, err := parseField(args[0])
	if err != nil {
		return nil, err
	}

	// 解析值
	value, err := parseValue(args[1])
	if err != nil {
		return nil, err
	}

	return []clause.Expression{clause.IN{Column: field, Values: []interface{}{value}}}, nil
}
