# 🚀 Query - Go语言查询构建器

一个简洁、高效的Go语言查询构建器，支持多种数据库操作，提供流畅的链式调用API。

## ✨ 功能特性

- ✅ **多操作支持** - SELECT、INSERT、UPDATE、DELETE多种SQL操作
- 🔗 **链式调用** - 流畅的API设计，支持链式调用
- 🔍 **复杂条件** - 支持AND、OR、NOT等复杂WHERE条件
- 📊 **排序分页** - 支持ORDER BY排序和LIMIT/OFFSET分页
- 🎯 **类型安全** - 支持多种数据类型的参数绑定
- 🪶 **轻量级** - 无外部依赖，轻量级设计
- 🌐 **AIP标准** - 支持Google AIP过滤和排序标准
- 🔌 **适配器架构** - 可扩展的适配器设计，支持不同数据库

## 📦 安装

```bash
go get github.com/epkgs/query
```

## 🚀 快速开始

### 🔍 SELECT 查询

```go
import "github.com/epkgs/query"

// 基本查询
q := query.Table("users").Select("id", "name", "age")

// 带 WHERE 条件的查询
q := query.Table("users").
    Where("age", ">", 18).
    Where("status", "active").
    Select("id", "name", "age")

// 流畅 API 写法（推荐）
q := query.Table("users").
    Eq("status", "active").
    Gte("age", 18).
    Lte("age", 65).
    Asc("created_at").
    Limit(10).
    Select("id", "name", "age")
```

### ➕ INSERT 操作

```go
// 单行插入（字段值对）
q := query.Table("users").Insert("name", "John", "age", 30)

// 使用 map 插入
q := query.Table("users").Insert(map[string]any{"name": "John", "age": 30})

// 批量插入
q := query.Table("users").Insert(
    map[string]any{"name": "John", "age": 30},
    map[string]any{"name": "Jane", "age": 25},
)
```

### ✏️ UPDATE 操作

```go
q := query.Table("users").Where("id", 1).Update("name", "John")

// 更新多个字段
q := query.Table("users").Where("id", 1).Update(map[string]any{
    "name": "John",
    "age":  30,
})
```

### 🗑️ DELETE 操作

```go
q := query.Table("users").Where("id", 1).Delete()
```

## 🔍 WHERE 条件

### 流畅 API（推荐）

```go
// 等于
q.Eq("name", "John")

// 不等于
q.Neq("status", "banned")

// 大于 / 大于等于
q.Gt("age", 18)
q.Gte("age", 18)

// 小于 / 小于等于
q.Lt("age", 65)
q.Lte("age", 65)

// LIKE
q.Like("name", "%John%")

// IN
q.In("id", 1, 2, 3)
q.In("id", []int{1, 2, 3})

// 逻辑组合
query.Or(
    query.Eq("status", "active"),
    query.Eq("role", "admin"),
)

query.And(
    query.Gte("age", 18),
    query.Lte("age", 65),
)

query.Not(query.Eq("status", "deleted"))
```

### 传统 API（兼容旧版本）

```go
q.Where("name", "=", "John")
q.Where("age", ">", 18)
q.Where("status", "IN", []string{"active", "pending"})
q.Where("email", nil)  // IS NULL
q.Where("age", ">", 18).OrWhere("role", "admin")
q.Not("status", "banned")
```

## 📊 ORDER BY 排序

```go
// 流畅 API
q.Asc("name")
q.Desc("age")
q.Asc("name").Desc("created_at")

// 传统 API
q.OrderBy("name")
q.OrderBy("age", "desc")
q.OrderBy("age desc, name asc")

// 结构化形式
q.OrderBy(clause.OrderBy{Column: "name", Desc: false})
```

## 📄 分页

```go
// LIMIT 和 OFFSET
q.Limit(10).Offset(20)

// 基于页码的分页（页码从 1 开始）
q.Paginate(3, 10)  // 第3页，每页10条
```

## 🔌 适配器架构

Query 库采用适配器模式，核心包只构建抽象查询表达式。通过适配器，查询可以转换为不同 ORM 或数据源的查询条件。

```
┌─────────────┐     ┌───────────────────┐     ┌──────────────────┐
│  AIP Filter │ ──→ │  clause.Where     │ ──→ │ GORM / Ent / SQL │
│  AIP OrderBy│     │  clause.OrderBys  │     │ 查询              │
└─────────────┘     └───────────────────┘     └──────────────────┘
```

### 🌐 AIP 适配器

支持 Google API Improvement Proposals (AIP) 的过滤和排序语法转换。

```go
import (
    query "github.com/epkgs/query"
    aip "github.com/epkgs/query/adapter/aip"
    filtering "go.einride.tech/aip/filtering"
    ordering "go.einride.tech/aip/ordering"
)

// 定义可过滤的字段声明
declarations, _ := filtering.NewDeclarations(
    filtering.DeclareFunc("name", "string", func(ctx context.Context, value string, match func(filtering.Filter) error) error {
        // ...
        return nil
    }),
)

// 解析 AIP 过滤条件 (如 filter=name="John" AND age>=18)
filter, err := filtering.ParseFilter(filteringRequest, declarations)
whereClause, err := aip.FromFilter(filter)

// 解析 AIP 排序条件 (如 order_by=name desc, age asc)
parsed, _ := ordering.ParseOrderBy(orderingRequest)
orderBys := aip.FromOrderBy(parsed)
```

支持的 AIP 过滤运算符：`=` `!=` `>` `>=` `<` `<=` `IN` `NOT` `AND` `OR`

### 🐬 GORM 适配器

将查询转换为 GORM Scope 函数，可与 db.Scopes() 配合使用。

```go
import (
    query "github.com/epkgs/query"
    gormadapter "github.com/epkgs/query/adapter/gorm"
    aip "github.com/epkgs/query/adapter/aip"
    "gorm.io/gorm"
)

// 基本用法
q := query.Table("users").
    Eq("status", "active").
    Gte("age", 18).
    Asc("created_at").
    Limit(10).
    Select("id", "name", "age")

db.Scopes(
    gormadapter.WhereScope(q.WhereExpr()),
    gormadapter.OrderByScope(q.OrderByExpr()),
    gormadapter.PaginationScope(q.PaginationExpr()),
).Find(&users)

// 组合使用
db.Scopes(gormadapter.QueryScope(
    q.WhereExpr(), q.OrderByExpr(), q.PaginationExpr(),
)).Find(&users)

// AIP 到 GORM 的完整示例
filter, _ := filtering.ParseFilter(request, declarations)
orderBy, _ := ordering.ParseOrderBy(orderingRequest)

whereClause, _ := aip.FromFilter(filter)
orderBys := aip.FromOrderBy(orderBy)

db.Scopes(gormadapter.Query(whereClause, orderBys, clause.Pagination{
    Limit:  &limit,
    Offset: offset,
})).Find(&users)
```

### 🔄 Ent 适配器

将查询转换为 Ent 的 sql.Selector 修改函数，支持字段名映射。

```go
import (
    query "github.com/epkgs/query"
    entadapter "github.com/epkgs/query/adapter/ent"
    aip "github.com/epkgs/query/adapter/aip"
    "entgo.io/ent/dialect/sql"
    "github.com/epkgs/query/clause"
)

// 基本用法
q := query.Table("users").
    Eq("status", "active").
    Desc("created_at").
    Select("id", "name")

// 使用 ExprHandler 映射字段名（Ent schema 列名可能与 API 字段名不同）
handler := func(expr clause.Expression) clause.Expression {
    switch e := expr.(type) {
    case clause.Eq:
        e.Col = fieldMappings[e.Col]
        return e
    case clause.Like:
        e.Col = fieldMappings[e.Col]
        return e
    }
    return expr
}

orderHandler := func(ob clause.OrderBy) clause.OrderBy {
    ob.Column = fieldMappings[ob.Column]
    return ob
}

// 应用到 Ent 客户端查询
client.User.Query().Modify(entadapter.Query(
    q.WhereExpr(),
    q.OrderByExpr(),
    clause.Pagination{Limit: &limit, Offset: offset},
    entadapter.WithExprHandler(handler),
    entadapter.WithOrderByHandler(orderHandler),
)).All(ctx)

// AIP 到 Ent 的完整示例
filter, _ := filtering.ParseFilter(request, declarations)
whereClause, _ := aip.FromFilter(filter)
orderBys := aip.FromOrderBy(parsedOrderBy)

client.User.Query().Modify(entadapter.Query(
    whereClause, orderBys, clause.Pagination{Limit: &limit},
    entadapter.WithExprHandler(columnMapper),
)).All(ctx)
```

### 📋 AIP → GORM/Ent 完整集成流程

以下是典型的 gRPC/gRPC-Gateway 服务中使用 AIP 过滤和排序的完整流程：

```go
func ListUsers(ctx context.Context, db *gorm.DB, req *pb.ListUsersRequest) ([]User, error) {
    var users []User

    // 1. 定义 AIP 过滤声明
    declarations, _ := filtering.NewDeclarations(
        filtering.DeclareString("name"),
        filtering.DeclareInt("age"),
        filtering.DeclareString("status"),
    )

    // 2. 解析请求中的 filter 和 order_by 参数
    filter, _ := filtering.ParseFilter(req.Filter, declarations)
    orderBy, _ := ordering.ParseOrderBy(req.OrderBy)

    // 3. 转换为 clause 格式
    whereClause, _ := aip.FromFilter(filter)
    orderBys := aip.FromOrderBy(orderBy)

    // 4. 应用分页
    pageSize := int(req.PageSize)
    offset := int((req.Page - 1) * req.PageSize)

    // 5. 使用 GORM 适配器执行查询
    db.Model(&User{}).Scopes(gormadapter.QueryScope(
        whereClause, orderBys, clause.Pagination{
            Limit:  &pageSize,
            Offset: offset,
        },
    )).Find(&users)

    return users, nil
}
```

## 🛠️ 构建 SQL

```go
// 自定义 Builder 实现
type MyBuilder struct {
    strings.Builder
    args []any
}

func (b *MyBuilder) WriteByte(c byte) error { b.Builder.WriteByte(c); return nil }
func (b *MyBuilder) AddVar(_ clause.Writer, vars ...any) {
    b.args = append(b.args, vars...)
    b.Builder.WriteString("?")
}
func (b *MyBuilder) WriteQuoted(field any) { b.Builder.WriteString(field.(string)) }
func (b *MyBuilder) AddError(err error) error { return err }

q := query.Table("users").Where("name", "John").Select("id", "name")
builder := &MyBuilder{}
q.Build(builder)
fmt.Println(builder.String()) // SELECT id, name FROM users WHERE name = ?
fmt.Println(builder.args)     // [John]
```

## ⚠️ 错误处理

```go
q := query.Table("users").Where("invalid", "value").Select("id")
if q.Error != nil {
    // 处理构建过程中的错误
    log.Fatal(q.Error)
}
```

## 📂 项目结构

```
query/
├── clause/          # 底层抽象组件（Expression, Where, OrderBy, Pagination）
├── adapter/
│   ├── aip/         # AIP 过滤和排序适配器
│   ├── gorm/        # GORM 适配器
│   └── ent/         # Ent 适配器
├── examples/
│   ├── aip-to-gorm/ # AIP → GORM 端到端示例
│   └── aip-to-ent/  # AIP → Ent 端到端示例
├── query.go         # 核心 Query 类型和入口函数
├── query_select.go  # SELECT 查询结构
├── query_insert.go  # INSERT 查询结构
├── query_update.go  # UPDATE 查询结构
├── query_delete.go  # DELETE 查询结构
├── component_*.go   # 可复用组件（where, orderbys, pagination）
└── query_test.go    # 测试文件
```

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test -v github.com/epkgs/query
go test -v github.com/epkgs/query/adapter/aip
```

## 📖 示例

查看 `examples` 目录下的完整端到端示例：

- `examples/aip-to-gorm/` - AIP 过滤 → clause → GORM 的完整集成示例
- `examples/aip-to-ent/` - AIP 过滤 → clause → Ent 的完整集成示例

## 许可证

MIT

## 贡献

欢迎提交 Issue 和 Pull Request！