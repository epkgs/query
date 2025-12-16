module github.com/epkgs/query/adapter/aip

go 1.23.0

require (
	github.com/epkgs/query v0.0.0-00010101000000-000000000000
	go.einride.tech/aip v0.76.0
	google.golang.org/genproto/googleapis/api v0.0.0-20250528174236-200df99c418a
)

require (
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250528174236-200df99c418a // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/epkgs/query => ../../
