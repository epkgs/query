package query

type chained[P any, V any] struct {
	Parent P
	Value  V
}

type Clauser[P any] interface {
	Where(field any, args ...any) P
	OrWhere(field any, args ...any) P
	Not(field any, args ...any) P
}

type Paginater[P any] interface {
	Pagination() Pagination
	Limit(limit int) P
	Offset(offset int) P
	Paginate(page int, pageSize int) P
}

type Orderer[P any] interface {
	OrderBy(column string, direction ...string) P
}
