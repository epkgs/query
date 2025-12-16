package main

import (
	"fmt"
	"log"

	"entgo.io/ent/dialect/sql"
	"github.com/epkgs/query/adapter/aip"
	entAdapter "github.com/epkgs/query/adapter/ent"
	"github.com/epkgs/query/clause"
	filtering "go.einride.tech/aip/filtering"
	ordering "go.einride.tech/aip/ordering"
)

// AIPRequest 实现了 filtering.Request、ordering.Request 接口
type AIPRequest struct {
	filter    string
	order     string
	pageSize  int32
	pageToken string
}

// GetFilter 返回过滤字符串
func (r AIPRequest) GetFilter() string {
	return r.filter
}

// GetOrderBy 返回排序字符串
func (r AIPRequest) GetOrderBy() string {
	return r.order
}

// GetPageSize 返回分页大小
func (r AIPRequest) GetPageSize() int32 {
	return r.pageSize
}

// GetPageToken 返回分页令牌
func (r AIPRequest) GetPageToken() string {
	return r.pageToken
}

// paginationRequest 定义分页请求接口
type paginationRequest interface {
	GetPageSize() int32
	GetPageToken() string
}

// parsePagination 将分页请求转换为 clause.Pagination
func parsePagination(pageReq paginationRequest) clause.Pagination {
	pageSize := int(pageReq.GetPageSize())
	if pageSize <= 0 {
		pageSize = 10 // 默认值
	}

	// 简单实现：将 pageToken 转换为 offset
	offset := 0

	return clause.Pagination{
		Limit:  &pageSize,
		Offset: offset,
	}
}

// createFilterDeclarations 创建过滤字段声明
func createFilterDeclarations() *filtering.Declarations {
	fmt.Println("Step 1: Creating filter declarations...")
	declarations, err := filtering.NewDeclarations(
		filtering.DeclareIdent("name", filtering.TypeString),
		filtering.DeclareIdent("age", filtering.TypeInt),
		filtering.DeclareIdent("status", filtering.TypeString),
		filtering.DeclareStandardFunctions(),
	)
	if err != nil {
		log.Fatalf("Error creating filter declarations: %v", err)
	}
	fmt.Println("✓ Filter declarations created successfully")
	fmt.Println()
	return declarations
}

// parseAIPRequests 解析 AIP 请求
func parseAIPRequests(req AIPRequest, filterDecls *filtering.Declarations) (filtering.Filter, ordering.OrderBy) {
	fmt.Println("Step 2: Parsing AIP requests...")

	// 解析过滤条件
	filter, err := filtering.ParseFilter(req, filterDecls)
	if err != nil {
		log.Fatalf("Error parsing filter: %v", err)
	}
	fmt.Println("✓ Filter parsed successfully")

	// 解析排序条件
	orderBy, err := ordering.ParseOrderBy(req)
	if err != nil {
		log.Fatalf("Error parsing order by: %v", err)
	}
	fmt.Println("✓ Order by parsed successfully")
	fmt.Println()

	return filter, orderBy
}

// convertToClauseObjects 将 AIP 请求转换为 clause 对象
func convertToClauseObjects(filter filtering.Filter, orderBy ordering.OrderBy, pageReq paginationRequest) (clause.Where, clause.OrderBys, clause.Pagination) {
	fmt.Println("Step 3: Converting to clause objects...")

	// 转换过滤条件
	whereClause, err := aip.FromFilter(filter)
	if err != nil {
		log.Fatalf("Error converting filter: %v", err)
	}
	fmt.Println("✓ Filter converted to clause.Where")

	// 转换排序条件
	orderBys := aip.FromOrderBy(orderBy)
	fmt.Println("✓ Order by converted to clause.OrderBys")

	// 创建分页条件
	paginationClause := parsePagination(pageReq)
	fmt.Println("✓ Pagination created as clause.Pagination")
	fmt.Println()

	return whereClause, orderBys, paginationClause
}

// convertToEntScopes 将 clause 对象转换为 Ent 查询函数
func convertToEntScopes(whereClause clause.Where, orderBys clause.OrderBys, paginationClause clause.Pagination) (func(*sql.Selector), func(*sql.Selector), func(*sql.Selector)) {
	fmt.Println("Step 4: Converting to Ent query functions...")

	// 转换过滤条件
	whereScope := entAdapter.Where(whereClause)
	fmt.Println("✓ Filter converted to Ent Where function")

	// 转换排序条件
	orderScope := entAdapter.OrderBy(orderBys)
	fmt.Println("✓ Order by converted to Ent OrderBy function")

	// 转换分页条件
	paginationScope := entAdapter.Pagination(paginationClause)
	fmt.Println("✓ Pagination converted to Ent Pagination function")
	fmt.Println()

	return whereScope, orderScope, paginationScope
}

// createEntSelector 创建 Ent Selector 并应用所有查询条件
func createEntSelector(whereScope, orderScope, paginationScope func(*sql.Selector)) *sql.Selector {
	fmt.Println("Step 5: Creating Ent Selector and applying scopes...")

	// 创建 Selector
	selector := sql.Selector{}
	// 使用 From 方法指定表名
	selector.From(sql.Table("users"))

	// 应用所有查询条件
	whereScope(&selector)
	orderScope(&selector)
	paginationScope(&selector)

	fmt.Println("✓ Ent Selector created and scopes applied")
	fmt.Println()

	return &selector
}

// printSQL 打印生成的 SQL
func printSQL(selector *sql.Selector) {
	// 打印 SQL
	fmt.Println("=== Generated SQL ===")

	// 获取生成的 SQL 查询
	query, args := selector.Query()
	fmt.Printf("SQL: %s\n", query)
	fmt.Printf("Args: %v\n", args)
}

func main() {
	// 创建 AIP 请求
	req := AIPRequest{
		filter:   "name = 'John' AND age > 18",
		order:    "created_at desc, name asc",
		pageSize: 10,
	}

	fmt.Println("=== AIP to Ent Example ===")
	fmt.Printf("AIP Filter: %s\n", req.GetFilter())
	fmt.Printf("AIP OrderBy: %s\n", req.GetOrderBy())
	fmt.Printf("AIP PageSize: %d\n\n", req.GetPageSize())

	// 1. 创建过滤字段声明
	filterDecls := createFilterDeclarations()

	// 2. 解析 AIP 请求
	filter, orderBy := parseAIPRequests(req, filterDecls)

	// 3. 转换为 clause 对象
	whereClause, orderBys, paginationClause := convertToClauseObjects(filter, orderBy, req)

	// 4. 转换为 Ent 查询函数
	whereScope, orderScope, paginationScope := convertToEntScopes(whereClause, orderBys, paginationClause)

	// 5. 创建 Ent Selector 并应用所有查询条件
	selector := createEntSelector(whereScope, orderScope, paginationScope)

	// 6. 打印生成的 SQL
	printSQL(selector)
}
