package main

import (
	"fmt"
	"log"

	"github.com/epkgs/query/adapter/aip"
	qgorm "github.com/epkgs/query/adapter/gorm"
	"github.com/epkgs/query/clause"
	filtering "go.einride.tech/aip/filtering"
	ordering "go.einride.tech/aip/ordering"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// AIPRequest 实现了 filtering.Request、ordering.Request 和 pagination.Request 接口
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

type paginationRequest interface {
	GetPageSize() int32
	GetPageToken() string
}

// parsePagination 将 pagination.Request 转换为 clause.Pagination
func parsePagination(pageReq paginationRequest) clause.Pagination {
	pageSize := int(pageReq.GetPageSize())

	// 将 pageToken 转换为 offset
	// 这里直接指定 offset 为 0
	// 实际项目需要根据 pagination token 解析 offset
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

// convertToGormScopes 将 clause 对象转换为 GORM scopes
func convertToGormScopes(whereClause clause.Where, orderBys clause.OrderBys, paginationClause clause.Pagination) (func(*gorm.DB) *gorm.DB, func(*gorm.DB) *gorm.DB, func(*gorm.DB) *gorm.DB) {
	fmt.Println("Step 4: Converting to GORM scopes...")

	// 转换过滤条件
	whereScope := qgorm.Where(whereClause)
	fmt.Println("✓ Filter converted to GORM where scope")

	// 转换排序条件
	orderScope := qgorm.OrderBy(orderBys)
	fmt.Println("✓ Order by converted to GORM order scope")

	// 转换分页条件
	paginationScope := qgorm.Pagination(paginationClause)
	fmt.Println("✓ Pagination converted to GORM pagination scope")
	fmt.Println()

	return whereScope, orderScope, paginationScope
}

// createGormDB 创建 GORM DB 实例
func createGormDB(whereScope, orderScope, paginationScope func(*gorm.DB) *gorm.DB) *gorm.DB {
	fmt.Println("Step 5: Creating GORM DB and applying scopes...")

	// 使用内存 SQLite 数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// 应用所有 scopes
	db = db.Scopes(whereScope, orderScope, paginationScope)
	fmt.Println("✓ GORM DB instance created")

	return db
}

// printSQL 打印生成的 SQL
func printSQL(db *gorm.DB) {
	// 打印 SQL
	fmt.Println("\n=== Generated SQL ===")

	// 使用 GORM 的 ToSQL 方法获取 SQL
	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		data := map[string]any{}
		return tx.Table("users").Find(&data)
	})
	fmt.Printf("SQL: %s\n", sql)
}

func main() {
	// 创建 AIP 请求
	req := AIPRequest{
		filter:   "name = 'John' AND age > 18",
		order:    "created_at desc, name asc",
		pageSize: 10,
	}

	fmt.Println("=== AIP to GORM Example ===")
	fmt.Printf("AIP Filter: %s\n", req.GetFilter())
	fmt.Printf("AIP OrderBy: %s\n", req.GetOrderBy())
	fmt.Printf("AIP PageSize: %d\n\n", req.GetPageSize())

	// 1. 创建过滤字段声明
	filterDecls := createFilterDeclarations()

	// 2. 解析 AIP 请求
	filter, orderBy := parseAIPRequests(req, filterDecls)

	// 3. 转换为 clause 对象
	whereClause, orderBys, paginationClause := convertToClauseObjects(filter, orderBy, req)

	// 4. 转换为 GORM scopes
	whereScope, orderScope, paginationScope := convertToGormScopes(whereClause, orderBys, paginationClause)

	// 5. 创建 GORM DB 实例
	db := createGormDB(whereScope, orderScope, paginationScope)

	// 6. 打印生成的 SQL
	printSQL(db)
}
