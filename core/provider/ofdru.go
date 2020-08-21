package provider

import (
	"Pixel/core/model"
	"Pixel/helper"
	"github.com/PharmaSpace/ofdru"
	"github.com/patrickmn/go-cache"
	"log"
	"strconv"
	"strings"
	"time"
)

type OfdRu struct {
	Cache    *cache.Cache
	Type     string
	Inn      string
	Login    string
	Password string
}

func (ofd *OfdRu) CheckReceipt(productName string, fd string, datePay time.Time, totalPrice int) (document model.Document, err error) {
	if receipts, ok := ofd.Cache.Get(productName); ok {
		for _, v := range receipts.([]ofdru.Receipt) {
			if fd == v.FD || fd == v.FP {
				document.Link = v.Link
				document.TotalSum = v.Price
			}
		}
	}
	return document, err
}

func (ofd *OfdRu) GetReceipts(date time.Time) {
	receipts, err := ofdru.OfdRu(ofd.Inn, ofd.Login, ofd.Password, "https://ofd.ru").GetReceipts(date)
	if err != nil {
		log.Printf("Ошибка получения чеков из ОФД: %s", receipts)
	}
	rCache := make(map[string][]model.Document)
	for _, receipt := range receipts {
		for _, product := range receipt.Products {
			product.Name = strings.ToLower(strings.Trim(product.Name, "\t \n"))
			name := helper.Cut(product.Name, 32)
			document := convertOfdruReceiptToDocument(receipt, product)
			rCache[name] = append(rCache[name], document)
		}

	}

	for k, _ := range rCache {
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

func (ofd *OfdRu) GetName() string {
	return ofd.Type
}

func convertOfdruReceiptToDocument(receipt ofdru.Receipt, product ofdru.Product) model.Document {
	date, _ := time.Parse("2016-07-26T12:32:00", receipt.Date)
	fd, _ := strconv.ParseInt(receipt.FD, 10, 32)
	return model.Document{
		DateTime:              date.Unix(),
		FiscalDocumentNumber:  int(fd),
		KktRegId:              receipt.KktRegId,
		Nds20:                 product.Vat,
		TotalSum:              receipt.Price,
		ProductName:           product.Name,
		ProductQuantity:       product.Quantity,
		ProductPrice:          product.Price,
		ProductTotalPrice:     product.TotalPrice,
		Link:                  receipt.Link,
		Ofd:                   "ofdru",
		FiscalDocumentNumber2: receipt.FP,
		FiscalDocumentNumber3: "",
	}
}
