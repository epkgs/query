module github.com/epkgs/query/adapter/ent

go 1.18.0

require (
	entgo.io/ent v0.11.0
	github.com/epkgs/query v0.0.0-00010101000000-000000000000
)

require github.com/google/uuid v1.3.0 // indirect

replace github.com/epkgs/query => ../../
