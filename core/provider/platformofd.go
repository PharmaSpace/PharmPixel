package provider

import (
	"github.com/PharmaSpace/platformOfd"
	"github.com/patrickmn/go-cache"
	"pixel/core/model"
	"pixel/helper"
	"strconv"
	"strings"
	"time"
)

// PlatformOfd структура
type PlatformOfd struct {
	Cache    *cache.Cache
	Type     string
	Login    string
	Password string
}

// CheckReceipt проверка чека
func (ofd *PlatformOfd) CheckReceipt(productName, fd string, datePay time.Time, totalPrice int) (document model.Document, err error) {
	if receipts, ok := ofd.Cache.Get(productName); ok {
		for _, v := range receipts.([]platformOfd.Receipt) {
			if fd == v.FD {
				document.Link = v.Link
				document.TotalSum = v.Price
			}
		}
	}
	return document, err
}

// GetReceipts получить чек
func (ofd *PlatformOfd) GetReceipts(date time.Time) {
	pOfd := platformOfd.PlatformOfd(ofd.Login, ofd.Password)
	receipts, _ := pOfd.GetReceipts(date)

	rCache := make(map[string][]model.Document)
	for _, receipt := range receipts {
		for _, product := range receipt.Products {
			document := convertPlatformOfdToDocument(receipt, product)
			name := helper.Cut(document.ProductName, 32)
			rCache[name] = append(rCache[name], document)
		}
	}

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

// GetName получит тип
func (ofd *PlatformOfd) GetName() string {
	return ofd.Type
}

func convertPlatformOfdToDocument(receipt platformOfd.Receipt, product platformOfd.Product) model.Document {
	var (
		document model.Document
	)
	date, _ := strconv.ParseInt(receipt.Date, 10, 64)
	date /= 1000
	fd, _ := strconv.ParseInt(receipt.FD, 10, 32)
	tp, _ := strconv.Atoi(strings.ReplaceAll(product.TotalPrice, ",", ""))
	document = model.Document{
		DateTime:              date,
		FiscalDocumentNumber:  int(fd),
		KktRegID:              "",
		Nds20:                 product.Vat,
		TotalSum:              receipt.Price,
		ProductName:           product.Name,
		ProductTotalPrice:     tp,
		Link:                  receipt.Link,
		Ofd:                   "platformOfd",
		FiscalDocumentNumber2: receipt.FP,
		FiscalDocumentNumber3: "",
	}
	for _, rp := range receipt.Products {
		if rp.Name == product.Name {
			document.ProductPrice = rp.Price
			document.ProductQuantity = rp.Quantity
			continue
		}
	}
	return document
}
