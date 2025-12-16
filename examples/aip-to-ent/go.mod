module aip-to-ent-example

go 1.23.0

require (
	entgo.io/ent v0.11.0
	github.com/epkgs/query v0.0.0
	github.com/epkgs/query/adapter/aip v0.0.0
	github.com/epkgs/query/adapter/ent v0.0.0
	go.einride.tech/aip v0.76.0
)

require (
	github.com/google/uuid v1.6.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250528174236-200df99c418a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250528174236-200df99c418a // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/epkgs/query => ../../

replace github.com/epkgs/query/adapter/aip => ../../adapter/aip

replace github.com/epkgs/query/adapter/ent => ../../adapter/ent
