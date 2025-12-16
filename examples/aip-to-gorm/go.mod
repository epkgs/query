module aip-to-gorm-example

go 1.23.0

replace github.com/epkgs/query => ../../

replace github.com/epkgs/query/adapter/aip => ../../adapter/aip

replace github.com/epkgs/query/adapter/gorm => ../../adapter/gorm

require (
	github.com/epkgs/query v0.0.0
	github.com/epkgs/query/adapter/aip v0.0.0
	github.com/epkgs/query/adapter/gorm v0.0.0
	go.einride.tech/aip v0.76.0
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250528174236-200df99c418a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250528174236-200df99c418a // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
