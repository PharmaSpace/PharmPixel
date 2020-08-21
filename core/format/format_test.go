package format

import (
	"Pixel/core/format/mock"
	"Pixel/core/model"
	"Pixel/store"
	"Pixel/store/service"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type formatMock struct {
}

func (f formatMock) Parse() {
}

func (f formatMock) GetCache(key string) (interface{}, bool) {
	if key == "пектусин таб. x10" {
		return document1, true
	}
	return nil, false
}

func (f formatMock) GetMP() service.MarketPlaceInterface {
	return mock.MarketPlaceMock{}
}

func (f formatMock) GetOFDCache(key string) (interface{}, bool) {
	if key == "пектусинтаб.х10" {
		return []service.MatchProductItem{matchProduct}, true
	}
	return nil, false
}

func (f formatMock) GetERPCache(key string) (interface{}, bool) {
	panic("implement me")
}

func (f formatMock) SetOFDCache(key string, val interface{}, duration time.Duration) {
	panic("implement me")
}

func (f formatMock) SetERPCache(key string, val interface{}, duration time.Duration) {
	panic("implement me")
}

func (f formatMock) GetDate() time.Time {
	return time.Now()
}

func setCache(localchache *cache.Cache, ofdItems map[string][]model.Document) {
	for k, v := range ofdItems {
		localchache.Set(k, v, 12*time.Hour)
	}
}

var document1 = []model.Document{{
	Ofd:                   "taxcom",
	DateTime:              1583058720,
	FiscalDocumentNumber:  92975,
	KktRegId:              "0002313753047155  ",
	Nds20:                 3899,
	TotalSum:              34080,
	ProductName:           "ПЕКТУСИН  ТАБ. Х10",
	ProductQuantity:       2,
	ProductPrice:          4100,
	ProductTotalPrice:     8200,
	Link:                  "https://receipt.taxcom.ru/v01/show?id=F2B2B9D1-8C85-424E-8E82-0F3E424A688C",
	FiscalDocumentNumber2: "MQQs+2pb",
	FiscalDocumentNumber3: "21",
}}
var document2 = 	 []model.Document{{
Ofd:                   "taxcom",
DateTime:              1583081940,
FiscalDocumentNumber:  111950,
KktRegId:              "0002315628026783  ",
Nds20:                 0,
TotalSum:              162700,
ProductName:           "2/5уп РУМАЛОН Р-Р В/М ВВЕД. АМП.",
ProductQuantity:       1,
ProductPrice:          117800,
ProductTotalPrice:     117800,
Link:                  "https://receipt.taxcom.ru/v01/show?id=9400B94A-1021-49B1-B528-47C1E399351F",
FiscalDocumentNumber2: "MQQMc3SC",
FiscalDocumentNumber3: "137",
}}

var ofdItems = map[string][]model.Document{
	"пектусин таб. x10": document1,
	"2/5уп румалон р-р в/м введ. амп.": document2,
}

var receipts = []service.Product{{
	Name:          "ПЕКТУСИН  ТАБ. Х10",
	Manufacturer:  "Random",
	Export:        true,
	PartNumber:    "124",
	Serial:        "13234343434",
	Stock:         1,
	WarehouseName: "warehouse1",
	SupplerName:   "OOO",
	SupplerInn:    "111111111111",
	CreatedAt:     "1583081940",
}}

var storeReciept = []store.Receipt{{
	DateTime:             "2020-03-01T13:32:00+03:00",
	FiscalDocumentNumber: 92975,
	Inn:                  "",
	KktRegId:             "0002313753047155  ",
	Link:                 "https://receipt.taxcom.ru/v01/show?id=F2B2B9D1-8C85-424E-8E82-0F3E424A688C",
	Name:                 "ПЕКТУСИН  ТАБ. Х10",
	Ofd:                  "taxcom",
	Price:                4100,
	PriceSellIn:          3899,
	ProductId:            "matchId",
	Quantity:             "2",
	Total:                0,
	TotalSum:             34080,
	SupplerName:          "sup",
	SupplerInn: 		"111111111111",
	PointName:            "5dbc2b02787b89a2305a49ee",
	Series:               "13234343434",
	IsValidated:          true,
	CreatedAt:            "2020-07-17T14:33:49.4270901+03:00",
	UpdatedAt:            "2020-07-17T14:33:49.4270901+03:00",
}}

var erpReceipts = []*model.Receipt{{
	PharmacyID:      "5dbc2b02787b89a2305a49ee",
	PharmacyAddress: "",
	Date:            "2020-07-17T14:33:49.4270901+03:00",
	KKM:             "",
	InvoiceNumber:   "92975",
	Manufacturer:    "",
	Supplier:        "sup",
	SupplierINN:     "111111111111",
	Name:            "пектусин таб. x10",
	PriceWoVat:      "3899",
	PriceWVat:       "4100",
	Vat:             "",
	TotalPrice:      "4100",
	TotalNumber:     "",
	ShipmentNumber:  "",
	Series:          "13234343434",
}}

var matchProduct = service.MatchProductItem{
	Export:       true,
	ID:           "matchId",
	Name:         "пектусин таб. x10",
	Stock:        3,
	SupplierName: "sup",
	SupplierInn:  "111111111111",
}


func Test_nameMatching (t *testing.T) {
	cache := cache.New(5*time.Minute, 10*time.Minute)
	setCache(cache, ofdItems)
	validReceipts, checkOfdProductNames := nameMatching(formatMock{}, erpReceipts, cache.Items())
	validReceipts[0].CreatedAt = storeReciept[0].CreatedAt
	validReceipts[0].UpdatedAt = storeReciept[0].UpdatedAt
	assert.Equal(t, storeReciept, validReceipts)
	assert.Equal(t, []string{}, checkOfdProductNames)
}
