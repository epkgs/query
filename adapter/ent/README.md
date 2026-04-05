# Ent 适配器

将 `github.com/epkgs/query` 查询构建器转换为 [Ent](https://entgo.io) ORM 框架的 SQL Selector 条件。

## 📦 安装

```bash
go get github.com/epkgs/query/adapter/ent
```

## 🚀 快速开始

### 基本用法

```go
package main

import (
    "context"
    
    "entgo.io/ent/dialect/sql"
    "github.com/epkgs/query"
    adapter "github.com/epkgs/query/adapter/ent"
    
    // 假设你的 ent schema 在这里
    "your-project/ent"
    "your-project/ent/user"
)

func main() {
    // 初始化 Ent 客户端
    client, err := ent.Open("mysql", "user:pass@tcp(localhost:3306)/dbname")
    if err != nil {
        panic(err)
    }
    defer client.Close()

    // 构建查询条件
    q := query.Table("users").
        Where("age", ">", 18).
        Where("city", "New York").
        OrderBy("name").
        Limit(10).
        Offset(20)

    // 方法一：使用 Query 组合函数
    ctx := context.Background()
    users, err := client.User.Query().
        Modify(adapter.Query(
            q.CloneWhereExpr(),
            q.CloneOrderByExpr(),
            q.ClonePaginationExpr(),
        )).
        All(ctx)

    // 方法二：分别应用各个组件
    users, err = client.User.Query().
        Modify(
            adapter.Where(q.CloneWhereExpr()),
            adapter.OrderBy(q.CloneOrderByExpr()),
            adapter.Pagination(q.ClonePaginationExpr()),
        ).
        All(ctx)
}
```

### 使用 SQL Selector

```go
import (
    "entgo.io/ent/dialect/sql"
    adapter "github.com/epkgs/query/adapter/ent"
)

// 构建查询条件
q := query.Where("name", "John").Where("age", ">", 18)

// 创建 SQL Selector
selector := sql.Select("*").From(sql.Table("users"))

// 应用条件
whereFunc := adapter.Where(q.CloneWhereExpr())
whereFunc(selector)

// 获取 SQL 和参数
sqlStr, args := selector.Query()
// 输出: SELECT * FROM `users` WHERE `name` = ? AND `age` > ?
// 参数: ["John", 18]
```

## 📚 API 文档

### Where

将 `clause.Where` 转换为 Ent 的 WHERE 条件修改函数。

```go
func Where(where clause.Where, opts ...Option) func(s *sql.Selector)
```

**示例：**

```go
q := query.Where("name", "John").Where("age", ">", 18)
whereClause := q.CloneWhereExpr()

client.User.Query().
    Modify(adapter.Where(whereClause)).
    All(ctx)
```

### OrderBy

将 `clause.OrderBys` 转换为 Ent 的 ORDER BY 修改函数。

```go
func OrderBy(orders clause.OrderBys, opts ...Option) func(s *sql.Selector)
```

**示例：**

```go
q := query.Table("users").OrderBy("age", "desc").OrderBy("name")
orderBys := q.CloneOrderByExpr()

client.User.Query().
    Modify(adapter.OrderBy(orderBys)).
    All(ctx)
```

### Pagination

将 `clause.Pagination` 转换为 Ent 的 LIMIT/OFFSET 修改函数。

```go
func Pagination(pagination clause.Pagination) func(s *sql.Selector)
```

**示例：**

```go
q := query.Table("users").Limit(10).Offset(20)
pagination := q.ClonePaginationExpr()

client.User.Query().
    Modify(adapter.Pagination(pagination)).
    All(ctx)
```

### Query

组合 WHERE、ORDER BY 和 PAGINATION 三个条件。

```go
func Query(where clause.Where, orders clause.OrderBys, pagination clause.Pagination, opts ...Option) func(s *sql.Selector)
```

**示例：**

```go
q := query.Table("users").
    Where("age", ">", 18).
    OrderBy("name").
    Limit(10)

client.User.Query().
    Modify(adapter.Query(
        q.CloneWhereExpr(),
        q.CloneOrderByExpr(),
        q.ClonePaginationExpr(),
    )).
    All(ctx)
```

## 🔧 高级特性

### ExprHandler - 表达式处理器

在转换为 Ent Predicate 前预处理 `clause.Expression`，可用于自定义字段映射、条件过滤等。

```go
type ExprHandler func(expr clause.Expression) clause.Expression
```

**示例：字段名映射**

```go
handler := func(expr clause.Expression) clause.Expression {
    switch e := expr.(type) {
    case clause.Eq:
        // 将 API 中的 user_name 映射到数据库的 name
        if e.Column == "user_name" {
            e.Column = "name"
        }
        return e
    case clause.Gt:
        // 将 API 中的 user_age 映射到数据库的 age
        if e.Column == "user_age" {
            e.Column = "age"
        }
        return e
    }
    return expr
}

client.User.Query().
    Modify(adapter.Where(whereClause, adapter.WithExprHandler(handler))).
    All(ctx)
```

**示例：条件过滤**

```go
handler := func(expr clause.Expression) clause.Expression {
    switch e := expr.(type) {
    case clause.Eq:
        // 过滤掉某些内部字段
        if e.Column == "internal_field" {
            return nil
        }
    }
    return expr
}

client.User.Query().
    Modify(adapter.Where(whereClause, adapter.WithExprHandler(handler))).
    All(ctx)
```

**示例：条件转换**

```go
// 将某些条件转换为其他形式
handler := func(expr clause.Expression) clause.Expression {
    switch e := expr.(type) {
    case clause.Eq:
        // 将等值查询转换为模糊查询
        if e.Column == "search" {
            return clause.Like{
                Column: "name",
                Value:  "%" + e.Value.(string) + "%",
            }
        }
    }
    return expr
}
```

### OrderByHandler - 排序处理器

在转换为 Ent 排序函数前预处理 `clause.OrderBy`，可用于字段映射、自定义排序逻辑等。

```go
type OrderHandler func(expr clause.OrderBy) clause.OrderBy
```

**示例：字段名映射**

```go
handler := func(order clause.OrderBy) clause.OrderBy {
    // 将 API 中的 user_name 映射到数据库的 name
    if order.Column == "user_name" {
        order.Column = "name"
    }
    return order
}

client.User.Query().
    Modify(adapter.OrderBy(orderBys, adapter.WithOrderByHandler(handler))).
    All(ctx)
```

**示例：默认排序**

```go
// 为某些字段添加默认排序方向
handler := func(order clause.OrderBy) clause.OrderBy {
    // created_at 默认降序
    if order.Column == "created_at" && !order.Desc {
        order.Desc = true
    }
    return order
}
```

## 🎯 支持的操作符

Ent 适配器支持以下查询操作符：

| 操作符 | 说明 | Ent 函数 |
|--------|------|----------|
| `=` | 等于 | `sql.EQ` |
| `!=`, `<>` | 不等于 | `sql.NEQ` |
| `>` | 大于 | `sql.GT` |
| `>=` | 大于等于 | `sql.GTE` |
| `<` | 小于 | `sql.LT` |
| `<=` | 小于等于 | `sql.LTE` |
| `LIKE` | 模糊匹配 | `sql.Like` |
| `IN` | 在集合中 | `sql.In` |

### 逻辑操作符

| 操作符 | 说明 | Ent 函数 |
|--------|------|----------|
| `AND` | 与（默认） | `sql.And` |
| `OR` | 或 | `sql.Or` |
| `NOT` | 非 | `sql.Not` |

## 💡 使用示例

### 复杂 WHERE 条件

```go
// (name = 'John' AND age > 18) OR (city = 'New York' AND age < 65)
q := query.Table("users").
    OrWhere(func(w query.Wherer) query.Wherer {
        w.Where("name", "John")
        w.Where("age", ">", 18)
        return w
    }).
    OrWhere(func(w query.Wherer) query.Wherer {
        w.Where("city", "New York")
        w.Where("age", "<", 65)
        return w
    })

users, err := client.User.Query().
    Modify(adapter.Where(q.CloneWhereExpr())).
    All(ctx)

// 生成 SQL: 
// SELECT * FROM `users` WHERE (`name` = ? AND `age` > ?) OR (`city` = ? AND `age` < ?)
```

### NOT 条件

```go
// NOT (name = 'John') AND NOT (age > 65)
q := query.Table("users").
    Not("name", "John").
    Not("age", ">", 65)

users, err := client.User.Query().
    Modify(adapter.Where(q.CloneWhereExpr())).
    All(ctx)

// 生成 SQL:
// SELECT * FROM `users` WHERE (NOT (`name` = ?)) AND (NOT (`age` > ?))
```

### IN 条件

```go
q := query.Where("id", "IN", []interface{}{1, 2, 3, 4, 5})

users, err := client.User.Query().
    Modify(adapter.Where(q.CloneWhereExpr())).
    All(ctx)

// 生成 SQL: 
// SELECT * FROM `users` WHERE `id` IN (?, ?, ?, ?, ?)
```

### LIKE 条件

```go
q := query.Where("name", "LIKE", "%John%")

users, err := client.User.Query().
    Modify(adapter.Where(q.CloneWhereExpr())).
    All(ctx)

// 生成 SQL: 
// SELECT * FROM `users` WHERE `name` LIKE ?
```

### 混合条件

```go
// WHERE city = 'New York' OR (name = 'John' AND NOT (age > 65))
q := query.Table("users").
    Where("city", "New York").
    OrWhere(func(w query.Wherer) query.Wherer {
        w.Where("name", "John")
        w.Not("age", ">", 65)
        return w
    })

users, err := client.User.Query().
    Modify(adapter.Where(q.CloneWhereExpr())).
    All(ctx)

// 生成 SQL:
// SELECT * FROM `users` WHERE `city` = ? OR (`name` = ? AND (NOT (`age` > ?)))
```

### 完整示例

```go
// WHERE + ORDER BY + PAGINATION + Handler
q := query.Table("users").
    Where("age", ">", 18).
    Where("city", "New York").
    OrWhere("role", "admin").
    OrderBy("age", "desc").
    OrderBy("name").
    Limit(10).
    Offset(20)

// 字段映射 handler
exprHandler := func(expr clause.Expression) clause.Expression {
    switch e := expr.(type) {
    case clause.Eq:
        // API 字段 -> 数据库字段
        mapping := map[string]string{
            "user_name": "name",
            "user_age":  "age",
        }
        if dbField, ok := mapping[e.Column]; ok {
            e.Column = dbField
        }
        return e
    }
    return expr
}

orderHandler := func(order clause.OrderBy) clause.OrderBy {
    // API 字段 -> 数据库字段
    mapping := map[string]string{
        "user_name": "name",
        "user_age":  "age",
    }
    if dbField, ok := mapping[order.Column]; ok {
        order.Column = dbField
    }
    return order
}

// 执行查询
users, err := client.User.Query().
    Modify(adapter.Query(
        q.CloneWhereExpr(),
        q.CloneOrderByExpr(),
        q.ClonePaginationExpr(),
        adapter.WithExprHandler(exprHandler),
        adapter.WithOrderByHandler(orderHandler),
    )).
    All(ctx)
```

### 与 Ent 原生条件结合

```go
// 可以将适配器与 Ent 原生条件结合使用
q := query.Where("age", ">", 18)

users, err := client.User.Query().
    Where(user.StatusEQ("active")).  // Ent 原生条件
    Modify(adapter.Where(q.CloneWhereExpr())). // 适配器条件
    All(ctx)
```

### 动态查询

```go
// 根据用户输入动态构建查询
func SearchUsers(nameFilter string, ageMin int, sortBy string) ([]*ent.User, error) {
    q := query.Table("users")
    
    if nameFilter != "" {
        q = q.Where("name", "LIKE", "%"+nameFilter+"%")
    }
    
    if ageMin > 0 {
        q = q.Where("age", ">=", ageMin)
    }
    
    if sortBy != "" {
        q = q.OrderBy(sortBy)
    }
    
    return client.User.Query().
        Modify(adapter.Query(
            q.CloneWhereExpr(),
            q.CloneOrderByExpr(),
            q.ClonePaginationExpr(),
        )).
        All(ctx)
}
```

## 🧪 测试

运行测试：

```bash
cd adapter/ent
go test -v
```

运行基准测试：

```bash
go test -bench=. -benchmem
```

## 📝 注意事项

1. **LIKE 值类型**: LIKE 操作符的值必须是字符串类型，否则会返回错误
2. **NOT 条件**: Ent 会保持 NOT 的语义（不像 GORM 会转换为反向操作符）
3. **错误处理**: 如果转换过程中发生错误，会通过 `selector.Err()` 返回
4. **NULL 值**: Ent 会自动处理 NULL 值的查询
5. **Modify vs Where**: 使用 `Modify` 可以直接操作底层的 SQL Selector，提供更大的灵活性

## 🔄 与 Ent Query 的集成

### 基本查询

```go
// 使用适配器
client.User.Query().
    Modify(adapter.Where(whereClause)).
    All(ctx)
```

### 聚合查询

```go
// Count
count, err := client.User.Query().
    Modify(adapter.Where(whereClause)).
    Count(ctx)

// Exist
exist, err := client.User.Query().
    Modify(adapter.Where(whereClause)).
    Exist(ctx)
```

### 部分字段查询

```go
var results []struct {
    Name string
    Age  int
}

err := client.User.Query().
    Modify(func(s *sql.Selector) {
        s.Select("name", "age")
        adapter.Where(whereClause)(s)
    }).
    Scan(ctx, &results)
```

### 分组查询

```go
var results []struct {
    City  string
    Count int
}

err := client.User.Query().
    Modify(func(s *sql.Selector) {
        s.Select("city", "COUNT(*) as count")
        adapter.Where(whereClause)(s)
        s.GroupBy("city")
    }).
    Scan(ctx, &results)
```

## 🆚 对比：Ent 原生 vs 适配器

| 功能 | Ent 原生 | 适配器 |
|------|----------|--------|
| 类型安全 | ✅ 编译时检查 | ❌ 运行时检查 |
| 动态条件 | ❌ 需要条件判断 | ✅ 统一的查询构建器 |
| 学习曲线 | 中等 | 低（统一 API） |
| 字段映射 | 手动处理 | ✅ Handler 支持 |
| 跨 ORM | ❌ 仅 Ent | ✅ 支持多种 ORM |

**推荐使用场景：**
- 使用 Ent 原生：需要类型安全、复杂的关联查询
- 使用适配器：动态查询、API 过滤、需要跨 ORM 兼容

## 🔗 相关链接

- [Query 主仓库](https://github.com/epkgs/query)
- [Ent 文档](https://entgo.io)
- [示例代码](../../examples/ent)

## 📄 许可证

MIT 许可证
