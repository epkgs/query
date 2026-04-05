# GORM 适配器

将 `github.com/epkgs/query` 查询构建器转换为 [GORM](https://gorm.io) ORM 框架的查询条件。

## 📦 安装

```bash
go get github.com/epkgs/query/adapter/gorm
```

## 🚀 快速开始

### 基本用法

```go
package main

import (
    "github.com/epkgs/query"
    adapter "github.com/epkgs/query/adapter/gorm"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

type User struct {
    ID   int
    Name string
    Age  int
    City string
}

func main() {
    // 初始化 GORM
    db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    if err != nil {
        panic(err)
    }

    // 构建查询条件
    q := query.Table("users").
        Where("age", ">", 18).
        Where("city", "New York").
        OrderBy("name").
        Limit(10).
        Offset(20)

    // 转换为 GORM scope 并执行查询
    var users []User
    err = db.Model(&User{}).
        Scopes(adapter.Query(
            q.CloneWhereExpr(),
            q.CloneOrderByExpr(),
            q.ClonePaginationExpr(),
        )).
        Find(&users).Error
}
```

## 📚 API 文档

### Where

将 `clause.Where` 转换为 GORM 的 WHERE 条件。

```go
func Where(where clause.Where, opts ...Option) func(db *gorm.DB) *gorm.DB
```

**示例：**

```go
q := query.Where("name", "John").Where("age", ">", 18)
whereClause := q.CloneWhereExpr()

db.Model(&User{}).
    Scopes(adapter.Where(whereClause)).
    Find(&users)
```

### OrderBy

将 `clause.OrderBys` 转换为 GORM 的 ORDER BY 子句。

```go
func OrderBy(orders clause.OrderBys, opts ...Option) func(db *gorm.DB) *gorm.DB
```

**示例：**

```go
q := query.Table("users").OrderBy("age", "desc").OrderBy("name")
orderBys := q.CloneOrderByExpr()

db.Model(&User{}).
    Scopes(adapter.OrderBy(orderBys)).
    Find(&users)
```

### Pagination

将 `clause.Pagination` 转换为 GORM 的 LIMIT/OFFSET 子句。

```go
func Pagination(pagination clause.Pagination) func(db *gorm.DB) *gorm.DB
```

**示例：**

```go
q := query.Table("users").Limit(10).Offset(20)
pagination := q.ClonePaginationExpr()

db.Model(&User{}).
    Scopes(adapter.Pagination(pagination)).
    Find(&users)
```

### Query

组合 WHERE、ORDER BY 和 PAGINATION 三个条件。

```go
func Query(where clause.Where, orders clause.OrderBys, pagination clause.Pagination, opts ...Option) func(db *gorm.DB) *gorm.DB
```

**示例：**

```go
q := query.Table("users").
    Where("age", ">", 18).
    OrderBy("name").
    Limit(10)

db.Model(&User{}).
    Scopes(adapter.Query(
        q.CloneWhereExpr(),
        q.CloneOrderByExpr(),
        q.ClonePaginationExpr(),
    )).
    Find(&users)
```

## 🔧 高级特性

### ExprHandler - 表达式处理器

在转换为 GORM 表达式前预处理 `clause.Expression`，可用于自定义字段映射、条件过滤等。

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
    }
    return expr
}

db.Model(&User{}).
    Scopes(adapter.Where(whereClause, adapter.WithExprHandler(handler))).
    Find(&users)
```

**示例：条件过滤**

```go
handler := func(expr clause.Expression) clause.Expression {
    switch e := expr.(type) {
    case clause.Eq:
        // 过滤掉某些字段
        if e.Column == "internal_field" {
            return nil
        }
    }
    return expr
}

db.Model(&User{}).
    Scopes(adapter.Where(whereClause, adapter.WithExprHandler(handler))).
    Find(&users)
```

### OrderByHandler - 排序处理器

在转换为 GORM 排序表达式前预处理 `clause.OrderBy`，可用于字段映射、自定义排序逻辑等。

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

db.Model(&User{}).
    Scopes(adapter.OrderBy(orderBys, adapter.WithOrderByHandler(handler))).
    Find(&users)
```

## 🎯 支持的操作符

GORM 适配器支持以下查询操作符：

| 操作符 | 说明 | GORM 表达式 |
|--------|------|-------------|
| `=` | 等于 | `gormClause.Eq` |
| `!=`, `<>` | 不等于 | `gormClause.Neq` |
| `>` | 大于 | `gormClause.Gt` |
| `>=` | 大于等于 | `gormClause.Gte` |
| `<` | 小于 | `gormClause.Lt` |
| `<=` | 小于等于 | `gormClause.Lte` |
| `LIKE` | 模糊匹配 | `gormClause.Like` |
| `IN` | 在集合中 | `gormClause.IN` |

### 逻辑操作符

| 操作符 | 说明 | GORM 表达式 |
|--------|------|-------------|
| `AND` | 与（默认） | `gormClause.And` |
| `OR` | 或 | `gormClause.Or` |
| `NOT` | 非 | `gormClause.Not` |

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

db.Model(&User{}).
    Scopes(adapter.Where(q.CloneWhereExpr())).
    Find(&users)
```

### NOT 条件

```go
// NOT (name = 'John') AND NOT (age > 65)
q := query.Table("users").
    Not("name", "John").
    Not("age", ">", 65)

db.Model(&User{}).
    Scopes(adapter.Where(q.CloneWhereExpr())).
    Find(&users)

// GORM 会将 NOT 转换为反向操作符：
// name <> 'John' AND age <= 65
```

### IN 条件

```go
q := query.Where("id", "IN", []interface{}{1, 2, 3, 4, 5})

db.Model(&User{}).
    Scopes(adapter.Where(q.CloneWhereExpr())).
    Find(&users)

// 生成: WHERE `id` IN (?,?,?,?,?)
```

### LIKE 条件

```go
q := query.Where("name", "LIKE", "%John%")

db.Model(&User{}).
    Scopes(adapter.Where(q.CloneWhereExpr())).
    Find(&users)

// 生成: WHERE `name` LIKE ?
```

### 组合使用

```go
// 完整示例：WHERE + ORDER BY + PAGINATION + Handler
q := query.Table("users").
    Where("age", ">", 18).
    Where("city", "New York").
    OrderBy("age", "desc").
    OrderBy("name").
    Limit(10).
    Offset(20)

// 字段映射 handler
exprHandler := func(expr clause.Expression) clause.Expression {
    switch e := expr.(type) {
    case clause.Eq:
        if e.Column == "user_name" {
            e.Column = "name"
        }
        return e
    }
    return expr
}

orderHandler := func(order clause.OrderBy) clause.OrderBy {
    if order.Column == "user_name" {
        order.Column = "name"
    }
    return order
}

// 执行查询
var users []User
err := db.Model(&User{}).
    Scopes(adapter.Query(
        q.CloneWhereExpr(),
        q.CloneOrderByExpr(),
        q.ClonePaginationExpr(),
        adapter.WithExprHandler(exprHandler),
        adapter.WithOrderByHandler(orderHandler),
    )).
    Find(&users).Error
```

## 🧪 测试

运行测试：

```bash
cd adapter/gorm
go test -v
```

运行基准测试：

```bash
go test -bench=. -benchmem
```

## 📝 注意事项

1. **NULL 值处理**: GORM 会自动处理 NULL 值的查询
2. **NOT 条件**: GORM 会将 NOT 条件转换为反向操作符（例如：NOT (age > 18) → age <= 18）
3. **作用域链**: 可以组合多个 scope 函数，GORM 会按顺序应用
4. **DryRun 模式**: 可以使用 DryRun 模式查看生成的 SQL 而不实际执行

## 🔗 相关链接

- [Query 主仓库](https://github.com/epkgs/query)
- [GORM 文档](https://gorm.io)
- [示例代码](../../examples/gorm)

## 📄 许可证

MIT 许可证
