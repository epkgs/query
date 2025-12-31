module github.com/epkgs/query/adapter/gorm

go 1.18.0

require (
	github.com/epkgs/query v0.0.0-00010101000000-000000000000
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	golang.org/x/text v0.20.0 // indirect
)

replace github.com/epkgs/query => ../..
