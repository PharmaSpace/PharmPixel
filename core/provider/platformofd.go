package provider

import (
	"Pixel/core/model"
	"github.com/PharmaSpace/platformOfd"
	"github.com/patrickmn/go-cache"
	"log"
	"strings"
	"time"
)

type PlatformOfd struct {
	Cache    *cache.Cache
	Type     string
	Login    string
	Password string
}

func (ofd *PlatformOfd) CheckReceipt(productName string, fd string, datePay time.Time, totalPrice int) (document model.Document, err error) {
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
func (ofd *PlatformOfd) GetReceipts(date time.Time) {
	pOfd := platformOfd.PlatformOfd(ofd.Login, ofd.Password)
	receipts, _ := pOfd.GetReceipts(date)

	rCache := make(map[string][]platformOfd.Receipt)
	for _, v := range receipts {
		for _, pr := range v.Products {
			name := cut(strings.ToLower(strings.Trim(pr.Name, "\t \n")), 32)
			rCache[name] = append(rCache[name], v)
		}
	}
	for k, v := range rCache {
		ofd.Cache.Set(k, v, 12*time.Hour)
	}
	log.Printf("Получено чеков: %d", len(rCache))
}

func (ofd *PlatformOfd) GetName() string {
	return ofd.Type
}
