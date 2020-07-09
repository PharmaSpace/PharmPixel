package provider

import (
	"Pixel/core/model"
	"github.com/PharmaSpace/ofdru"
	"github.com/patrickmn/go-cache"
	"log"
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
	receipts, _ := ofdru.OfdRu(ofd.Inn, ofd.Login, ofd.Password, "https://ofd.ru").GetReceipts(date)
	rCache := make(map[string][]ofdru.Receipt)
	for _, v := range receipts {
		for _, pr := range v.Products {
			name := cut(strings.ToLower(strings.Trim(pr.Name, "\t \n")), 32)
			rCache[name] = append(rCache[name], v)
		}
	}
	for k, _ := range rCache {
		if item, ok := ofd.Cache.Get(k); ok {
			receipts := item.([]ofdru.Receipt)
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
