package query_test

import (
	"fmt"

	"github.com/epkgs/query"
)

// Example_select demonstrates a basic SELECT query with WHERE, ORDER BY and pagination.
func Example_select() {
	q := query.Table("users").
		Eq("status", "active").
		Gte("age", 18).
		Asc("created_at").
		Limit(10).
		Select("id", "name", "age")

	fmt.Println(q.Error == nil)
	// Output: true
}

// Example_selectWhere demonstrates SELECT with complex WHERE conditions.
func Example_selectWhere() {
	q := query.Table("users").
		Where("age", ">", 18).
		Where("status", "active").
		Select("id", "name", "email")

	fmt.Println(q.Error == nil)
	// Output: true
}

// Example_insert demonstrates INSERT queries using different input forms.
func Example_insert() {
	// Field-value pairs
	q1 := query.Table("users").Insert("name", "John", "age", 30)
	fmt.Println(q1.Error == nil)

	// Map input
	q2 := query.Table("users").Insert(map[string]any{"name": "Jane", "age": 25})
	fmt.Println(q2.Error == nil)

	// Batch insert
	q3 := query.Table("users").Insert(
		map[string]any{"name": "Alice", "age": 28},
		map[string]any{"name": "Bob", "age": 32},
	)
	fmt.Println(q3.Error == nil)

	// Output:
	// true
	// true
	// true
}

// Example_update demonstrates UPDATE queries.
func Example_update() {
	q := query.Table("users").
		Where("id", 1).
		Update("name", "John Updated")

	fmt.Println(q.Error == nil)
	// Output: true
}

// Example_delete demonstrates DELETE queries.
func Example_delete() {
	q := query.Table("users").
		Where("id", 1).
		Delete()

	fmt.Println(q.Error == nil)
	// Output: true
}

// Example_whereConditions demonstrates the fluent comparison operators.
func Example_whereConditions() {
	q := query.Table("products").
		Eq("status", "published").  // WHERE status = 'published'
		Gte("price", 100).          // AND price >= 100
		Lte("price", 500).          // AND price <= 500
		Neq("category", "deleted"). // AND category != 'deleted'
		Like("name", "%phone%").    // AND name LIKE '%phone%'
		In("tag", "new", "hot").    // AND tag IN ('new', 'hot')
		Select("id", "name", "price")

	fmt.Println(q.Error == nil)
	// Output: true
}

// Example_logicalCombination demonstrates AND, OR, NOT logical combinations.
func Example_logicalCombination() {
	// Build conditions on a base Query before converting to SelectQuery
	q := query.Table("users")

	// OR combination: admin users OR (active AND adult users)
	q.Or(
		query.Eq("role", "admin"),
		query.And(
			query.Eq("status", "active"),
			query.Gte("age", 18),
		),
	)

	// Exclude deleted users with NOT
	q.Not(query.Eq("status", "deleted"))

	// Convert to SELECT query
	sq := q.Select("id", "name")
	fmt.Println(sq.Error == nil)
	// Output: true
}

// Example_orderBy demonstrates ORDER BY sorting.
func Example_orderBy() {
	q := query.Table("users").
		Desc("created_at").   // ORDER BY created_at DESC
		Asc("name").          // , name ASC
		Select("id", "name")

	fmt.Println(q.Error == nil)
	// Output: true
}

// Example_pagination demonstrates pagination methods.
func Example_pagination() {
	// Using Limit/Offset
	q1 := query.Table("users").Limit(20).Offset(40).Select("id", "name")
	fmt.Println(q1.Error == nil)

	// Using Paginate (page 3, 20 per page = offset 40, limit 20)
	q2 := query.Table("users").Paginate(3, 20).Select("id", "name")
	fmt.Println(q2.Error == nil)

	// Output:
	// true
	// true
}

// Example_aipToGorm demonstrates the full AIP → clause → GORM integration flow.
// This is the most common pattern for gRPC/gRPC-Gateway services.
func Example_aipToGorm() {
	// In real code, parse filter and order_by from request parameters:
	// filter, _ := filtering.ParseFilter(req.Filter, declarations)
	// orderBy, _ := ordering.ParseOrderBy(req.OrderBy)

	// Convert AIP filter to clause.Where
	// whereClause, _ := aip.FromFilter(filter)

	// Convert AIP order_by to clause.OrderBys
	// orderBys := aip.FromOrderBy(orderBy)

	// Apply to GORM query:
	// db.Model(&User{}).Scopes(
	//     gormadapter.Where(whereClause),
	//     gormadapter.OrderBy(orderBys),
	//     gormadapter.Pagination(clause.Pagination{Limit: &limit, Offset: offset}),
	// ).Find(&users)

	// Alternatively, use the combined Query function:
	// db.Scopes(gormadapter.Query(whereClause, orderBys, clause.Pagination{Limit: &limit})).Find(&users)

	fmt.Println("AIP → GORM integration ready")
	// Output: AIP → GORM integration ready
}

// Example_aipToEnt demonstrates the full AIP → clause → Ent integration flow.
func Example_aipToEnt() {
	// In real code:
	// filter, _ := filtering.ParseFilter(req.Filter, declarations)
	// whereClause, _ := aip.FromFilter(filter)

	// orderBys := aip.FromOrderBy(parsedOrderBy)

	// Use ExprHandler to map API field names to Ent schema columns:
	// fieldMappings := map[string]string{"userName": "user_name", "createdAt": "created_at"}
	// handler := func(expr clause.Expression) clause.Expression {
	//     switch e := expr.(type) {
	//     case clause.Eq:
	//         e.Column = fieldMappings[e.Column]
	//         return e
	//     }
	//     return expr
	// }

	// Apply to Ent client:
	// client.User.Query().Modify(entadapter.Query(
	//     whereClause, orderBys, clause.Pagination{Limit: &limit},
	//     entadapter.WithExprHandler(handler),
	// )).All(ctx)

	fmt.Println("AIP → Ent integration ready")
	// Output: AIP → Ent integration ready
}
