package mock

import (
	"database/sql"
	"github.com/stretchr/testify/mock"
)

type DBMock struct {
	mock.Mock
}

func (d *DBMock) Close() error {
	return nil
}

func (d *DBMock) Query(query string, args ...interface{}) (*sql.Rows, error) {
	mockArgs := d.Called(query, args)
	result, ok := mockArgs.Get(0).(*sql.Rows)
	if ok {
		return result, mockArgs.Error(1)
	}
	return nil,  mockArgs.Error(1)
}


