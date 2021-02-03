package provider

import (
	"github.com/PharmaSpace/oneofd"
	"github.com/patrickmn/go-cache"
	"log"
	"pixel/core/model"
	"pixel/helper"
	"pixel/sentry"
	"strconv"
	"time"
)

// OneOfd структура
type OneOfd struct {
	Cache    *cache.Cache
	Type     string
	Login    string
	Password string
	Sentry   *sentry.Sentry
}

// CheckReceipt проверка чека
func (ofd *OneOfd) CheckReceipt(productName, fd string, datePay time.Time, totalPrice int) (document model.Document, err error) {
	if receipts, ok := ofd.Cache.Get(productName); ok {
		for _, v := range receipts.([]oneofd.Receipt) {
			if fd == v.FD || fd == v.FP {
				document.Link = v.Link
				document.TotalSum = v.Price
				document.KktRegID = v.KktRegId
			}
		}
	}
	return document, err
}

// GetReceipts получить чеки
func (ofd *OneOfd) GetReceipts(date time.Time) {
	receipts, err := oneofd.OneOfd(ofd.Login, ofd.Password).GetReceipts(date)
	if err != nil {
		ofd.Sentry.Error(err)
	}

	rCache := make(map[string][]model.Document)
	for _, v := range receipts {
		for _, pr := range v.Products {
			document := convertOneOfdToDocument(v, pr)
			name := helper.Cut(pr.Name, 32)
			rCache[name] = append(rCache[name], document)
		}
	}

	// Специальный кусок чтобы добавить в ключ чеков
	for k := range rCache {
		if item, ok := ofd.Cache.Get(k); ok {
			receipts := item.([]model.Document)
			rCache[k] = append(rCache[k], receipts...)
		}
	}

	for k, v := range rCache {
		ofd.Cache.Set(k, v, 12*time.Hour)
	}
	log.Printf("Получено чеков: %d", len(rCache))
}

func convertOneOfdToDocument(receipt oneofd.Receipt, product oneofd.Product) model.Document {
	document := model.Document{}

	date, _ := time.Parse("2006-01-02T15:04:05Z", receipt.Date)

	document.DateTime = date.Unix()

	document.Ofd = "1ofd"

	fiscalDocumentNumber, _ := strconv.Atoi(receipt.FD)
	document.KktRegID = receipt.KktRegId
	document.Link = receipt.Link
	document.TotalSum = receipt.Price
	document.FiscalDocumentNumber = fiscalDocumentNumber
	document.FiscalDocumentNumber2 = receipt.FP

	productPrice := product.Price
	if product.TotalPrice < productPrice {
		productPrice = product.TotalPrice
	}

	document.ProductPrice = productPrice
	document.ProductTotalPrice = product.TotalPrice
	document.ProductQuantity = product.Quantity
	if document.ProductQuantity == 0 {
		document.ProductQuantity = int(float64(product.TotalPrice / product.Price))
	}
	document.ProductName = product.Name

	return document
}

// GetName получение типа
func (ofd *OneOfd) GetName() string {
	return ofd.Type
}
