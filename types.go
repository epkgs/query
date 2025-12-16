package query

import "github.com/epkgs/query/clause"

type Wherer interface {
	WhereExpr() clause.Where
	Where(field any, args ...any) Wherer
	OrWhere(field any, args ...any) Wherer
	Not(field any, args ...any) Wherer
}

type genericWherer[Self any] interface {
	WhereExpr() clause.Where
	Where(field any, args ...any) Self
	OrWhere(field any, args ...any) Self
	Not(field any, args ...any) Self
}

type Paginator[Self any] interface {
	Pagination() Pagination
	Limit(limit int) Self
	Offset(offset int) Self
	Paginate(page int, pageSize int) Self
}

type Sorter[Self any] interface {
	OrderBy(column string, direction ...string) Self
}
