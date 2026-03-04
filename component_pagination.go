package query

import "github.com/epkgs/query/clause"

type genericPaginator[Q any] interface {
	PaginationExpr() clause.Pagination
	Limit(limit int) Q
	Offset(offset int) Q
	Paginate(page int, pageSize int) Q
}

type Pagination struct {
	Offset int
	Limit  int
}

var _ genericPaginator[*Query] = (*pagination[*Query])(nil)
var _ clause.Expression = (*pagination[*Query])(nil)

type pagination[Q any] struct {
	Parent Q
	Value  clause.Pagination
}

func (p *pagination[Q]) PaginationExpr() clause.Pagination {
	return p.Value
}

// Limit 设置查询的限制条数
func (p *pagination[Q]) Limit(limit int) Q {
	p.Value.Limit = &limit
	return p.Parent
}

// Offset 设置查询的偏移量
func (p *pagination[Q]) Offset(offset int) Q {
	p.Value.Offset = offset
	return p.Parent
}

// Paginate 设置分页参数
func (p *pagination[Q]) Paginate(page int, pageSize int) Q {
	if page > 0 {
		p.Offset((page - 1) * pageSize)
	}
	if pageSize > 0 {
		p.Limit(pageSize)
	}
	return p.Parent
}

func (p *pagination[Q]) Build(builder clause.Builder) {
	p.Value.Build(builder)
}
