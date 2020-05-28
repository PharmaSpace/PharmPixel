package provider

import (
	"Pixel/core/model"
	"github.com/PharmaSpace/taxcom"
	"github.com/patrickmn/go-cache"
	"strings"
	"time"
)

type TaxCom struct {
	Cache        *cache.Cache
	Type         string
	Login        string
	Password     string
	IdIntegrator string
}

func (ofd *TaxCom) CheckReceipt(productName string, fd string, datePay time.Time, totalPrice int) (document model.Document, err error) {
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

func (ofd *TaxCom) GetReceipts(date time.Time) {
	accountList := taxcom.Taxcom(ofd.Login, ofd.Password, ofd.IdIntegrator, "").GetAccountList()
	for _, account := range accountList {
		receipts, _ := taxcom.Taxcom(ofd.Login, ofd.Password, ofd.IdIntegrator, account).GetReceipts(date)
		rCache := make(map[string][]taxcom.Receipt)
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
}

func (ofd *TaxCom) GetName() string {
	return ofd.Type
}
