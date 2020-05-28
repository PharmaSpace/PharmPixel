package provider

import (
	"Pixel/core/model"
	"github.com/PharmaSpace/oneofd"
	"github.com/patrickmn/go-cache"
	"strings"
	"time"
)

type OneOfd struct {
	Cache    *cache.Cache
	Type     string
	Login    string
	Password string
}

func (ofd *OneOfd) CheckReceipt(productName string, fd string, datePay time.Time, totalPrice int) (document model.Document, err error) {
	if receipts, ok := ofd.Cache.Get(productName); ok {
		for _, v := range receipts.([]oneofd.Receipt) {
			if fd == v.FD || fd == v.FP {
				document.Link = v.Link
				document.TotalSum = v.Price
			}
		}
	}
	return document, err
}

func (ofd *OneOfd) GetReceipts(date time.Time) {
	receipts, _ := oneofd.OneOfd(ofd.Login, ofd.Password).GetReceipts(date)
	rCache := make(map[string][]oneofd.Receipt)
	for _, v := range receipts {
		for _, pr := range v.Products {
			name := cut(strings.ToLower(strings.Trim(pr.Name, "\t \n")), 32)
			rCache[name] = append(rCache[name], v)
		}
	}
	for k, v := range rCache {
		ofd.Cache.Set(k, v, 12*time.Hour)
	}
}

func (ofd *OneOfd) GetName() string {
	return ofd.Type
}