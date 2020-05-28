package engine

// Package engine defines interfaces each supported storage should implement.
// Includes default implementation with boltdb

import (
	"Pixel/store"
)

// NOTE: mockery works from linked to go-path and with GOFLAGS='-mod=vendor' go generate
//go:generate sh -c "mockery -inpkg -name Interface -print > /tmp/engine-mock.tmp && mv /tmp/engine-mock.tmp engine_mock.go"

// Interface defines methods provided by low-level storage engine
type Interface interface {
	Create(file store.File) (store.File, error) // create new comment, avoid dups by id
	Get(req GetRequest) (store.File, error)     // get comment by id

	CreateProduct(product store.Product) (store.Product, error)
	GetProduct(req GetRequest) (store.Product, error)

	CreateReceipt(receipt store.Receipt) (store.Receipt, error)
	CreateReceiptN(receipt store.ReceiptN) (store.ReceiptN, error)
	GetAllReceipt() ([]store.Receipt, error)
	DeleteReceipt(store.Receipt) error

	Close() error // close storage engine
}

// GetRequest is the input for Get func
type GetRequest struct {
	Name string `json:"name"`
}
