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
        Scopes(adapter.QueryScope(
            q.WhereExpr(),
            q.OrderByExpr(),
            q.PaginationExpr(),
        )).
        Find(&users).Error
}
```

## 📚 API 文档

### WhereScope

将 `clause.Where` 转换为 GORM 的 WHERE 条件 Scope 函数。

```go
func WhereScope(where clause.Where, convs ...WhereConverter) func(db *gorm.DB) *gorm.DB
```

**示例：**

```go
q := query.Where("name", "John").Where("age", ">", 18)
whereClause := q.WhereExpr()

db.Model(&User{}).
    Scopes(adapter.WhereScope(whereClause)).
    Find(&users)
```

### WhereExpr

将 `clause.Where` 转换为 GORM 的 `clause.Where` 表达式，可直接传入 `db.Where()` 或组合使用。

```go
func WhereExpr(where clause.Where, convs ...WhereConverter) gormClause.Expression
```

### OrderByScope

将 `clause.OrderBys` 转换为 GORM 的 ORDER BY 子句 Scope 函数。

```go
func OrderByScope(orders clause.OrderBys, convs ...OrderByConverter) func(db *gorm.DB) *gorm.DB
```

**示例：**

```go
q := query.Table("users").OrderBy("age", "desc").OrderBy("name")
orderBys := q.OrderByExpr()

db.Model(&User{}).
    Scopes(adapter.OrderByScope(orderBys)).
    Find(&users)
```

### OrderByExpr

将 `clause.OrderBys` 转换为 GORM 的 `clause.OrderBy` 表达式。

```go
func OrderByExpr(orders clause.OrderBys, convs ...OrderByConverter) gormClause.OrderBy
```

### PaginationScope

将 `clause.Pagination` 转换为 GORM 的 LIMIT/OFFSET 子句 Scope 函数。

```go
func PaginationScope(pagination clause.Pagination) func(db *gorm.DB) *gorm.DB
```

**示例：**

```go
q := query.Table("users").Limit(10).Offset(20)
pagination := q.PaginationExpr()

db.Model(&User{}).
    Scopes(adapter.PaginationScope(pagination)).
    Find(&users)
```

### QueryScope

组合 WHERE、ORDER BY 和 PAGINATION 三个条件，是 WhereScope、OrderByScope、PaginationScope 的便捷组合。

```go
func QueryScope(where clause.Where, orders clause.OrderBys, pagination clause.Pagination) func(db *gorm.DB) *gorm.DB
```

**示例：**

```go
q := query.Table("users").
    Where("age", ">", 18).
    OrderBy("name").
    Limit(10)

db.Model(&User{}).
    Scopes(adapter.QueryScope(
        q.WhereExpr(),
        q.OrderByExpr(),
        q.PaginationExpr(),
    )).
    Find(&users)
```

## 🔧 高级特性

### WhereConverter - 表达式转换器

在转换为 GORM 表达式时，可通过 `WhereConverter` 自定义转换逻辑。若转换成功（`converted` 为 `true`），使用自定义结果；否则由默认逻辑处理。可用于自定义字段映射、条件过滤等。

```go
type WhereConverter func(e clause.Expression) (gormExpr gormClause.Expression, converted bool)
```

**示例：字段名映射**

```go
columnMapper := func(e clause.Expression) (gormClause.Expression, bool) {
    cmp, ok := e.(clause.ComparisonExpression)
    if ok && cmp.Column() == "user_name" {
        // 将 API 中的 user_name 映射到数据库的 name
        return gormClause.Eq{Column: gormClause.Column{Name: "name"}, Value: cmp.Value()}, true
    }
    return nil, false
}

db.Model(&User{}).
    Scopes(adapter.WhereScope(whereClause, columnMapper)).
    Find(&users)
```

**示例：条件过滤**

```go
filter := func(e clause.Expression) (gormClause.Expression, bool) {
    if cmp, ok := e.(clause.ComparisonExpression); ok && cmp.Column() == "internal_field" {
        // 过滤掉 internal_field 条件，返回 nil 表示不生成任何 GORM 表达式
        return nil, true
    }
    return nil, false // 返回 false 表示由默认逻辑处理
}

db.Model(&User{}).
    Scopes(adapter.WhereScope(whereClause, filter)).
    Find(&users)
```

### OrderByConverter - 排序转换器

在转换为 GORM 排序表达式时，可通过 `OrderByConverter` 自定义转换逻辑。若转换成功（`converted` 为 `true`），使用自定义结果；否则由默认逻辑处理。可用于字段映射、自定义排序逻辑等。

```go
type OrderByConverter func(o clause.OrderBy) (gormOrder gormClause.OrderByColumn, converted bool)
```

**示例：字段名映射**

```go
orderMapper := func(o clause.OrderBy) (gormClause.OrderByColumn, bool) {
    // 将 API 中的 user_name 映射到数据库的 name
    if o.Column == "user_name" {
        return gormClause.OrderByColumn{
            Column: gormClause.Column{Name: "name"},
            Desc:   o.Desc,
        }, true
    }
    return gormClause.OrderByColumn{}, false
}

db.Model(&User{}).
    Scopes(adapter.OrderByScope(orderBys, orderMapper)).
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
    Scopes(adapter.WhereScope(q.WhereExpr())).
    Find(&users)
```

### NOT 条件

```go
// NOT (name = 'John') AND NOT (age > 65)
q := query.Table("users").
    Not("name", "John").
    Not("age", ">", 65)

db.Model(&User{}).
    Scopes(adapter.WhereScope(q.WhereExpr())).
    Find(&users)

// GORM 会将 NOT 转换为反向操作符：
// name <> 'John' AND age <= 65
```

### IN 条件

```go
q := query.Where("id", "IN", []interface{}{1, 2, 3, 4, 5})

db.Model(&User{}).
    Scopes(adapter.WhereScope(q.WhereExpr())).
    Find(&users)

// 生成: WHERE `id` IN (?,?,?,?,?)
```

### LIKE 条件

```go
q := query.Where("name", "LIKE", "%John%")

db.Model(&User{}).
    Scopes(adapter.WhereScope(q.WhereExpr())).
    Find(&users)

// 生成: WHERE `name` LIKE ?
```

### 组合使用

```go
// 完整示例：WHERE + ORDER BY + PAGINATION + Converter
q := query.Table("users").
    Where("age", ">", 18).
    Where("city", "New York").
    OrderBy("age", "desc").
    OrderBy("name").
    Limit(10).
    Offset(20)

// 字段映射 converter
columnMapper := func(e clause.Expression) (gormClause.Expression, bool) {
    if cmp, ok := e.(clause.ComparisonExpression); ok && cmp.Column() == "user_name" {
        return gormClause.Eq{Column: gormClause.Column{Name: "name"}, Value: cmp.Value()}, true
    }
    return nil, false
}

orderMapper := func(o clause.OrderBy) (gormClause.OrderByColumn, bool) {
    if o.Column == "user_name" {
        return gormClause.OrderByColumn{
            Column: gormClause.Column{Name: "name"},
            Desc:   o.Desc,
        }, true
    }
    return gormClause.OrderByColumn{}, false
}

// 执行查询
var users []User
err := db.Model(&User{}).
    Scopes(adapter.QueryScope(
        q.WhereExpr(),
        q.OrderByExpr(),
        q.PaginationExpr(),
    )).Find(&users).Error

// 或使用单独的 Scope 函数并传入 Converter：
err = db.Model(&User{}).
    Scopes(
        adapter.WhereScope(q.WhereExpr(), columnMapper),
        adapter.OrderByScope(q.OrderByExpr(), orderMapper),
        adapter.PaginationScope(q.PaginationExpr()),
    ).Find(&users).Error
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
