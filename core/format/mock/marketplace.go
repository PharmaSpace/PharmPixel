package mock

import (
	"Pixel/store"
	"Pixel/store/service"
	"github.com/stretchr/testify/mock"
)

type MarketPlaceMock struct {
	mock.Mock
}

func (m MarketPlaceMock) SendProduct(products []service.Product) {
}

func (m MarketPlaceMock) SendOfdProducts(products []service.Product, isOfd bool, isErp bool) {
	m.Called(products, isOfd, isErp)
}

func (m MarketPlaceMock) SendReceipt(receipts []store.Receipt) {
	m.Called(receipts)
}

func (m MarketPlaceMock) SendReceiptN(receipts []store.ReceiptN) {
}

func (m MarketPlaceMock) GetMatchProduct() service.MatchProducts {
	args := m.Called()
	return args.Get(0).(service.MatchProducts)
}

func (m MarketPlaceMock) GetMatchProducts(filterDate string, isOfd bool, isErp bool) service.MatchProducts {
	args := m.Called()
	return args.Get(0).(service.MatchProducts)
}

