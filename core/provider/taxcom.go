package provider

import (
	"fmt"
	"github.com/PharmaSpace/taxcom"
	"github.com/patrickmn/go-cache"
	"pixel/core/model"
	"pixel/helper"
	"strconv"
	"time"
)

// TaxCom структура
type TaxCom struct {
	Cache        *cache.Cache
	Type         string
	Login        string
	Password     string
	IDIntegrator string
}

// CheckReceipt  проверка чека
func (ofd *TaxCom) CheckReceipt(productName, fd string, datePay time.Time, totalPrice int) (document model.Document, err error) {
	if receipts, ok := ofd.Cache.Get(productName); ok {
		for _, v := range receipts.([]taxcom.Receipt) {
			if fd == v.FD || fd == v.FP {
				document.Link = v.Link
				document.TotalSum = v.Price
			}
		}
	}
	return document, err
}

// GetReceipts получить чека
func (ofd *TaxCom) GetReceipts(date time.Time) {
	accountList := taxcom.Taxcom(ofd.Login, ofd.Password, ofd.IDIntegrator, "").GetAccountList()
	for _, account := range accountList {
		receipts, _ := taxcom.Taxcom(ofd.Login, ofd.Password, ofd.IDIntegrator, account).GetReceipts(date)
		rCache := make(map[string][]model.Document)
		for _, v := range receipts {
			for _, pr := range v.Products {
				document := ofd.buildDocument(v, pr)
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
	}
}

func (ofd *TaxCom) buildDocument(receipt taxcom.Receipt, product taxcom.Product) model.Document {
	document := model.Document{}

	date, _ := time.Parse("2006-01-02T15:04:05Z", receipt.Date)

	document.DateTime = date.Unix()

	document.Ofd = "taxcom"

	fiscalDocumentNumber, _ := strconv.Atoi(receipt.FD)
	document.KktRegID = receipt.KktRegId
	document.Link = receipt.Link
	document.TotalSum = receipt.Price
	document.FiscalDocumentNumber = fiscalDocumentNumber
	document.FiscalDocumentNumber2 = receipt.FP
	document.FiscalDocumentNumber3 = fmt.Sprint(receipt.DocumentNumber)

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

// GetName получить тип
func (ofd *TaxCom) GetName() string {
	return ofd.Type
}
