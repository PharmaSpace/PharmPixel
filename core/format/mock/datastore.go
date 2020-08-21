package mock

import (
	"Pixel/store"
	"github.com/stretchr/testify/mock"
)

type DSMock struct {
	mock.Mock
}

func (d *DSMock) Create(file store.File) (fileName store.File, err error) {
	args := d.Called(file)
	return store.File{}, args.Error(1)
}

func (d *DSMock) CreateProduct(product store.Product) (pr store.Product, err error) {
	args := d.Called(product)
	return args.Get(0).(store.Product), args.Error(1)
}

func (d *DSMock) Get(fileName string) (store.File, error) {
	args := d.Called(fileName)
	return args.Get(0).(store.File), args.Error(1)
}

func (d *DSMock) GetProduct(name string) (store.Product, error) {
	args := d.Called(name)
	return args.Get(0).(store.Product), args.Error(1)
}

func (d *DSMock) CreateReceipt(receipt store.Receipt) (store.Receipt, error) {
	args := d.Called(receipt)
	return args.Get(0).(store.Receipt), args.Error(1)
}

func (d *DSMock) CreateReceiptN(receipt store.ReceiptN) (store.ReceiptN, error) {
	args := d.Called(receipt)
	return args.Get(0).(store.ReceiptN), args.Error(1)
}

func (d *DSMock) GetAllReceipts() ([]store.Receipt, error) {
	args := d.Called()
	return args.Get(0).([]store.Receipt), args.Error(1)
}

func (d *DSMock) DeleteReceipt(receipt store.Receipt) error {
	args := d.Called(receipt)
	return args.Error(0)
}

func (d *DSMock) Close() error {
	return nil
}



