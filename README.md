# query

一个简洁、高效的Go语言查询构建器，支持多种数据库操作，提供流畅的链式调用API。

## 功能特性

- 支持SELECT、INSERT、UPDATE、DELETE等多种SQL操作
- 流畅的链式调用API设计
- 支持复杂的WHERE条件（AND、OR、NOT）
- 支持排序（ORDER BY），支持多种输入类型
- 支持分页（LIMIT、OFFSET）
- 支持多种数据类型的参数绑定
- 轻量级设计，无外部依赖
- 支持AIP（API Improvement Proposals）过滤和排序标准
- 可扩展的适配器设计，支持不同的数据库

## 安装

```bash
go get github.com/epkgs/query
```

## 快速开始

### SELECT查询

```go
import "github.com/epkgs/query"

// 基本查询
q := query.Table("users").Select("id", "name", "age")

// 带WHERE条件
q := query.Table("users").Where("age", ">", 18).Select("id", "name", "age")

// 带排序
q := query.Table("users").OrderBy("age", "desc").OrderBy("name").Select("id", "name", "age")

// 带分页
q := query.Table("users").Limit(10).Offset(20).Select("id", "name", "age")

// 链式调用
q := query.Table("users").Where("age", ">", 18).OrWhere("name", "admin").OrderBy("age", "desc").Limit(10).Select("id", "name", "age")
```

### INSERT操作

```go
// 单行插入（使用字段值对）
q := query.Table("users").Insert("name", "John", "age", 30)

// 单行插入（使用map）
q := query.Table("users").Insert(map[string]interface{}{"name": "John", "age": 30})

// 多行插入
q := query.Table("users").Insert(
    map[string]interface{}{"name": "John", "age": 30},
    map[string]interface{}{"name": "Jane", "age": 25},
)
```

### UPDATE操作

```go
// 更新单个字段
q := query.Table("users").Where("id", 1).Update("name", "John")

// 更新多个字段（使用map）
q := query.Table("users").Where("id", 1).Update(map[string]interface{}{"name": "John", "age": 30})
```

### DELETE操作

```go
// 删除所有记录
q := query.Table("users").Delete()

// 带条件删除
q := query.Table("users").Where("id", 1).Delete()
```

## WHERE条件

### 基本条件

```go
// 等于
q.Where("name", "=", "John")
q.Where("name", "John") // 简写形式

// 不等于
q.Where("name", "<>", "John")
q.Not("name", "John") // 简写形式

// 大于
q.Where("age", ">", 18)

// 小于
q.Where("age", "<", 30)

// 大于等于
q.Where("age", ">=", 18)

// 小于等于
q.Where("age", "<=", 30)

// LIKE
q.Where("name", "LIKE", "J%")

// IN
q.Where("id", []interface{}{1, 2, 3, 4, 5})

// NULL
q.Where("email", nil)
```

### 逻辑组合

```go
// AND条件（默认）
q.Where("age", ">", 18).Where("age", "<", 30)

// OR条件
q.Where("age", ">", 18).OrWhere("name", "admin")

// NOT条件
q.Where("age", ">", 18).Not("status", "banned")
```

## ORDER BY排序

### 字符串形式

```go
// 单个字段升序
q.OrderBy("name")

// 单个字段降序
q.OrderBy("age", "desc")

// 多个字段排序
q.OrderBy("age", "desc").OrderBy("name")

// 逗号分隔的多个字段
q.OrderBy("age desc, name asc")
```

### 结构化形式

```go
import "github.com/epkgs/query/clause"

// 单个clause.OrderBy
q.OrderBy(clause.OrderBy{Column: "name", Desc: false})

// 多个clause.OrderBy参数
q.OrderBy(
    clause.OrderBy{Column: "age", Desc: true},
    clause.OrderBy{Column: "name", Desc: false},
)

// []clause.OrderBy切片
orderBys := []clause.OrderBy{
    {Column: "age", Desc: true},
    {Column: "name", Desc: false},
}
q.OrderBy(orderBys)

// clause.OrderBys集合
orderBys := clause.OrderBys{
    {Column: "age", Desc: true},
    {Column: "name", Desc: false},
}
q.OrderBy(orderBys)
```

## 分页

```go
// 基本分页
q.Limit(10).Offset(20)

// 使用Paginate方法（页码从1开始）
q.Paginate(3, 10) // 第3页，每页10条
```

## 构建SQL

查询构建完成后，可以通过`Build`方法将查询转换为SQL语句：

```go
type Builder interface {
    WriteString(s string)
    WriteQuoted(field interface{})
    AddVar(writer clause.Writer, vars ...interface{})
    AddError(err error)
}

// 自定义Builder实现
q := query.Table("users").Where("name", "John").Select("id", "name")
builder := &YourCustomBuilder{}
q.Build(builder)
```

## 适配器

该库设计了适配器机制，可以将查询转换为不同数据库的SQL语句。目前支持的适配器：

- GORM适配器
- Ent适配器

### GORM适配器

```go
import (
    adapter "github.com/epkgs/query/adapter/gorm"
    "gorm.io/gorm"
)

q := query.Table("users").Where("name", "John").Select("id", "name")
db, err := gorm.Open(...) // 初始化GORM
var users []User
db.Scopes(adapter.Where(q.WhereExpr())).Find(&users)
```

### Ent适配器

```go
import (
	"entgo.io/ent/dialect/sql"
    adapter "github.com/epkgs/query/adapter/ent"
)

q := query.Table("users").Where("name", "John").Select("id", "name")
// 使用 Selector 构建查询
selector := sql.Selector{}
selector.From(sql.Table("users"))
// 转换过滤条件
whereScope := adapter.Where(q.WhereExpr())
orderScope := adapter.OrderBy(q.OrderByExpr())
paginationScope := adapter.Pagination(q.PaginationExpr())
// 应用所有查询条件
whereScope(&selector)
orderScope(&selector)
paginationScope(&selector)
// 执行操作
query, args := selector.Query()
```

## 错误处理

查询构建过程中如果发生错误，会将错误赋值给 `Error` 属性。可以通过 `Error` 属性判断是否有错误发生：

```go
q := query.Table("users").Where("invalid_field", "value").Select("id", "name")
if err := q.Error; err != nil {
    // 处理错误
}
```

## 测试

运行测试：

```bash
go test ./...
```

## 示例

查看`examples`目录下的示例代码：

- `examples/gorm` - GORM适配器示例
- `examples/ent` - Ent适配器示例

## 许可证

MIT许可证

## 贡献

欢迎提交Issue和Pull Request！

## 联系方式

如有问题或建议，请提交Issue。