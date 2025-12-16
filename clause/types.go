package clause

type Valuer interface {
	Value() (any, error)
}
