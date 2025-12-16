package query

import "github.com/epkgs/query/clause"

type Pagination struct {
	Offset int
	Limit  int
}

type pagination[P any] struct {
	Parent P
	Value  clause.Pagination
}

func (p *pagination[P]) PaginationExpr() clause.Pagination {
	return p.Value
}

// Limit 设置查询的限制条数
func (p *pagination[P]) Limit(limit int) P {
	p.Value.Limit = &limit
	return p.Parent
}

// Offset 设置查询的偏移量
func (p *pagination[P]) Offset(offset int) P {
	p.Value.Offset = offset
	return p.Parent
}

// Paginate 设置分页参数
func (p *pagination[P]) Paginate(page int, pageSize int) P {
	if page > 0 {
		p.Offset((page - 1) * pageSize)
	}
	if pageSize > 0 {
		p.Limit(pageSize)
	}
	return p.Parent
}

func (p *pagination[P]) Build(builder clause.Builder) {
	p.Value.Build(builder)
}
