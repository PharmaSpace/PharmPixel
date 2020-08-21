package db

import "database/sql"

type DB interface {
	Close() error
	Query(query string, args ...interface{}) (*sql.Rows, error)
}