package provider

import (
	"github.com/PharmaSpace/sbis"
	"github.com/patrickmn/go-cache"
	"log"
	"pixel/core/model"
	"pixel/helper"
	"pixel/sentry"
	"strconv"
	"strings"
	"time"
)

// Sbis структура
type Sbis struct {
	Cache    *cache.Cache
	Type     string
	Inn      string
	Login    string
	Password string
	Sentry   *sentry.Sentry
}

// CheckReceipt проверка чека
func (ofd *Sbis) CheckReceipt(productName, fp string, datePay time.Time, totalPrice int) (document model.Document, err error) {
	if receipts, ok := ofd.Cache.Get(strings.ToUpper(productName)); ok {
		for _, v := range receipts.([]*sbis.Receipt) {
			if fp == strconv.Itoa(v.RequestNumber) || totalPrice == v.TotalSum {
				for _, i := range v.Items {
					if !strings.EqualFold(i.Name, productName) {
						continue
					}
					date, _ := time.Parse("2006-01-02T15:04:05", v.ReceiveDateTime)
					document.DateTime = date.Unix()
					document.Link = v.Url
					document.FiscalDocumentNumber = v.FiscalDocumentNumber
					document.KktRegID = v.KktRegID
					document.ProductPrice = i.Price
					document.TotalSum = v.TotalSum
				}
			}
		}
	}
	return document, err
}

// GetReceipts получить чек
func (ofd *Sbis) GetReceipts(date time.Time) {
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	endDate := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 59, time.Local)

	receipts, err := sbis.GetReceipts(ofd.Inn, startDate.Format("2006-01-02T15:04:05"), endDate.Format("2006-01-02T15:04:05"), sbis.SetAuthConfig(&sbis.AuthConfig{
		AppClientID: "1025293145607151",
		Login:       ofd.Login,
		Password:    ofd.Password,
	}))
	if err != nil {
		ofd.Sentry.Error(err)
	}

	rCache := make(map[string][]model.Document)
	for _, v := range receipts {
		for i, pr := range v.Items {
			name := helper.Cut(pr.Name, 32)
			document := sbisReceiptToDocument(v, i)
			rCache[name] = append(rCache[name], document)
		}
	}

	for k := range rCache {
		if item, ok := ofd.Cache.Get(k); ok {
			receipts := item.([]*sbis.Receipt)
			for _, receipt := range receipts {
				for i := range receipt.Items {
					rCache[k] = append(rCache[k], sbisReceiptToDocument(receipt, i))
				}
			}

		}
	}

	for k, v := range rCache {
		ofd.Cache.Set(k, v, 12*time.Hour)
	}
	log.Printf("Получено чеков: %d", len(rCache))
}

// GetName получить тип
func (ofd *Sbis) GetName() string {
	return ofd.Type
}

func sbisReceiptToDocument(receipt *sbis.Receipt, itemNum int) model.Document {
	document := model.Document{
		DateTime:              int64(receipt.DateTime),
		FiscalDocumentNumber:  receipt.FiscalDocumentNumber,
		KktRegID:              receipt.KktRegID,
		Nds20:                 receipt.NdsNo,
		TotalSum:              receipt.TotalSum,
		ProductName:           receipt.Items[itemNum].Name,
		ProductQuantity:       int(receipt.Items[itemNum].Quantity),
		ProductPrice:          receipt.Items[itemNum].Price,
		ProductTotalPrice:     receipt.TotalSum,
		Link:                  receipt.Url,
		Ofd:                   "sbis",
		FiscalDocumentNumber2: receipt.FiscalDriveNumber,
		FiscalDocumentNumber3: strconv.FormatInt(receipt.FiscalSign, 10),
	}
	return document
}
