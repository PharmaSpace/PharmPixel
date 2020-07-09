package provider

import (
	"Pixel/core/model"
	"github.com/PharmaSpace/sbis"
	"github.com/patrickmn/go-cache"
	"log"
	"strconv"
	"strings"
	"time"
)

type Sbis struct {
	Cache    *cache.Cache
	Type     string
	Inn      string
	Login    string
	Password string
}

func (ofd *Sbis) CheckReceipt(productName string, fp string, datePay time.Time, totalPrice int) (document model.Document, err error) {
	if receipts, ok := ofd.Cache.Get(strings.ToUpper(productName)); ok {
		for _, v := range receipts.([]*sbis.Receipt) {
			if fp == strconv.Itoa(v.RequestNumber) || totalPrice == v.TotalSum {
				for _, i := range v.Items {
					if i.Name == strings.ToUpper(productName) {
						date, _ := time.Parse("2006-01-02T15:04:05", v.ReceiveDateTime)
						document.DateTime = date.Unix()
						document.Link = v.Url
						document.FiscalDocumentNumber = v.FiscalDocumentNumber
						document.KktRegId = v.KktRegID
						document.ProductPrice = i.Price
						document.TotalSum = v.TotalSum
					}
				}
			}
		}
	}
	return document, err
}
func (ofd *Sbis) GetReceipts(date time.Time) {
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	endDate := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 59, time.Local)

	receipts, _ := sbis.GetReceipts(ofd.Inn, startDate.Format("2006-01-02T15:04:05"), endDate.Format("2006-01-02T15:04:05"), sbis.SetAuthConfig(&sbis.AuthConfig{
		AppClientID: "1025293145607151",
		Login:       ofd.Login,
		Password:    ofd.Password,
	}))

	rCache := make(map[string][]*sbis.Receipt)
	for _, v := range receipts {
		for _, pr := range v.Items {
			name := cut(strings.ToLower(strings.Trim(pr.Name, "\t \n")), 32)
			rCache[name] = append(rCache[name], v)
		}
	}

	for k, _ := range rCache {
		if item, ok := ofd.Cache.Get(k); ok {
			receipts := item.([]*sbis.Receipt)
			rCache[k] = append(rCache[k], receipts...)
		}
	}

	for k, v := range rCache {
		ofd.Cache.Set(k, v, 12*time.Hour)
	}
	log.Printf("Получено чеков: %d", len(rCache))
}

func (ofd *Sbis) GetName() string {
	return ofd.Type
}
