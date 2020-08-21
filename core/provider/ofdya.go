package provider

import (
	"Pixel/core/model"
	"Pixel/helper"
	"github.com/PharmaSpace/OfdYa"
	"github.com/patrickmn/go-cache"
	"log"
	"time"
)

type Ofdya struct {
	Cache *cache.Cache
	Type  string
	Token string
}

func (ofd *Ofdya) CheckReceipt(productName string, fd string, datePay time.Time, totalPrice int) (document model.Document, err error) {
	if receipts, ok := ofd.Cache.Get(productName); ok {
		for _, v := range receipts.([]OfdYa.Receipt) {
			if fd == v.FD || fd == v.FP {
				document.Link = v.Link
				document.TotalSum = v.Price
			}
		}
	}
	return document, err
}

func (ofd *Ofdya) GetReceipts(date time.Time) {
	pOfd := OfdYa.OfdYa(ofd.Token)
	receipts, _ := pOfd.GetReceipts(date)

	rCache := make(map[string][]OfdYa.Receipt)
	for _, v := range receipts {
		for _, pr := range v.Products {
			name := helper.Cut(pr.Name, 32)
			rCache[name] = append(rCache[name], v)
		}
	}

	for k, _ := range rCache {
		if item, ok := ofd.Cache.Get(k); ok {
			receipts := item.([]OfdYa.Receipt)
			rCache[k] = append(rCache[k], receipts...)
		}
	}

	for k, v := range rCache {
		ofd.Cache.Set(k, v, 12*time.Hour)
	}
	log.Printf("Получено чеков: %d", len(rCache))
}

func (ofd *Ofdya) GetName() string {
	return ofd.Type
}
