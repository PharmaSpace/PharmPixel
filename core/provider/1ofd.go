package provider

import (
	"github.com/PharmaSpace/oneofd"
	"github.com/patrickmn/go-cache"
	"log"
	"pixel/core/model"
	"pixel/helper"
	"time"
)

// OneOfd структура
type OneOfd struct {
	Cache    *cache.Cache
	Type     string
	Login    string
	Password string
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
	receipts, _ := oneofd.OneOfd(ofd.Login, ofd.Password).GetReceipts(date)
	rCache := make(map[string][]oneofd.Receipt)
	for _, v := range receipts {
		for _, pr := range v.Products {
			name := helper.Cut(pr.Name, 32)
			rCache[name] = append(rCache[name], v)
		}
	}

	for k := range rCache {
		if item, ok := ofd.Cache.Get(k); ok {
			receipts := item.([]oneofd.Receipt)
			rCache[k] = append(rCache[k], receipts...)
		}
	}

	for k, v := range rCache {
		ofd.Cache.Set(k, v, 12*time.Hour)
	}
	log.Printf("Получено чеков: %d", len(rCache))
}

// GetName получение типа
func (ofd *OneOfd) GetName() string {
	return ofd.Type
}
