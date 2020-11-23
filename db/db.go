package db

import "database/sql"

// DB интерфейс для интгарции с бд
type DB interface {
	Close() error
	Query(query string, args ...interface{}) (*sql.Rows, error)
}
