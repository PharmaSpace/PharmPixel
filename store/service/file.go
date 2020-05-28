package service

import (
	"Pixel/store"
	"Pixel/store/engine"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"time"
)

// DataStore wraps store.Interface with additional methods
type DataStore struct {
	Engine engine.Interface
}

// Create prepares comment and forward to Interface.Create
func (s *DataStore) Create(file store.File) (fileName store.File, err error) {
	if file, err = s.prepareNewFile(file); err != nil {
		return store.File{}, errors.Wrap(err, "failed to prepare file")
	}
	return s.Engine.Create(file)
}

// Create prepares comment and forward to Interface.Create
func (s *DataStore) CreateProduct(product store.Product) (pr store.Product, err error) {
	if product, err = s.prepareNewProduct(product); err != nil {
		return store.Product{}, errors.Wrap(err, "failed to prepare file")
	}
	_, err = s.Engine.CreateProduct(product)
	return pr, err
}

// Get file by Name
func (s *DataStore) Get(fileName string) (store.File, error) {
	c, err := s.Engine.Get(engine.GetRequest{Name: fileName})
	if err != nil {
		return store.File{}, err
	}
	return c, nil
}

// GeProduct file by Name
func (s *DataStore) GeProduct(name string) (store.Product, error) {
	c, err := s.Engine.GetProduct(engine.GetRequest{Name: name})
	if err != nil {
		return store.Product{}, err
	}
	return c, nil
}

// CreateReceipt
func (s *DataStore) CreateReceipt(receipt store.Receipt) (store.Receipt, error) {
	c, err := s.Engine.CreateReceipt(receipt)
	if err != nil {
		return store.Receipt{}, err
	}
	return c, nil
}

// CreateReceiptN
func (s *DataStore) CreateReceiptN(receipt store.ReceiptN) (store.ReceiptN, error) {
	c, err := s.Engine.CreateReceiptN(receipt)
	if err != nil {
		return store.ReceiptN{}, err
	}
	return c, nil
}

func (s *DataStore) GetAllReceipts() ([]store.Receipt, error) {
	c, err := s.Engine.GetAllReceipt()
	if err != nil {
		return []store.Receipt{}, err
	}
	return c, nil
}

func (s *DataStore) DeleteReceipt(receipt store.Receipt) error {
	return s.Engine.DeleteReceipt(receipt)
}

// Close store service
func (s *DataStore) Close() error {
	return s.Engine.Close()
}

// prepareNewComment sets new comment fields, hashing and sanitizing data
func (s *DataStore) prepareNewFile(file store.File) (store.File, error) {
	// fill ID and time if empty
	if file.ID == "" {
		file.ID = uuid.New().String()
	}
	if file.Timestamp.IsZero() {
		file.Timestamp = time.Now()
	}
	return file, nil
}

// prepareNewProduct sets new comment fields, hashing and sanitizing data
func (s *DataStore) prepareNewProduct(product store.Product) (store.Product, error) {
	// fill ID and time if empty
	if product.ID == "" {
		product.ID = uuid.New().String()
	}
	if product.Timestamp.IsZero() {
		product.Timestamp = time.Now()
	}
	return product, nil
}
