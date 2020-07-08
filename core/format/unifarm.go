package format

import (
	"Pixel/config"
	"Pixel/core/model"
	"Pixel/store"
	"Pixel/store/service"
	"github.com/PharmaSpace/platformOfd"
	serviceLib "github.com/kardianos/service"
	"github.com/patrickmn/go-cache"
	"strconv"
	"strings"
	"time"
)

type uniFarm struct {
	date        time.Time
	Config      *config.Config
	DataService *service.DataStore
	MP          *service.Marketpalce
	Log         serviceLib.Logger
	cache       *cache.Cache
}

func UniFarm(c *config.Config, dataService *service.DataStore, mp *service.Marketpalce, ca *cache.Cache, log serviceLib.Logger, date time.Time) *uniFarm {
	return &uniFarm{Config: c, DataService: dataService, MP: mp, Log: log, cache: ca, date: date}
}

func (u *uniFarm) Parse() {
	uni := service.UniFarm(u.Config.UniFarmOptions.Username, u.Config.UniFarmOptions.Password)
	productsRow := uni.GetProduct(u.date)
	products := []service.Product{}
	receiptsN := []store.ReceiptN{}
	receipts := []store.Receipt{}

	for _, product := range productsRow {
		products = append(products, service.Product{
			Name:         product.ProductName,
			Manufacturer: product.ManufacturerName,
			Stock:        0,
			PartNumber:   product.PartNumber,
			Serial:       product.Serial,
		})
	}
	u.MP.SendProduct(products)

	u.getMachProduct()
	//time.Sleep(10 * time.Minute)

	receiptsRow := uni.GetReceipt(u.date)
	for _, receipt := range receiptsRow {
		datePay, _ := time.Parse("2006-01-02T15:04:05", receipt.Date)

		totalPrice, err := strconv.ParseFloat(receipt.SumPriceSellOut, 32)
		if err != nil {
			u.Log.Errorf("Ошибка привидения суммы заказа к float: %s %v", receipt.ProductName, err.Error())
		}
		document, checkReceiptErr := u.checkReceipt(receipt.ProductName, receipt.NumberFiscalDocument, datePay, int(totalPrice))

		priceSellIn, err := strconv.ParseFloat(receipt.PriceSellIn, 32)
		priceSellOut, err := strconv.ParseFloat(receipt.PriceSellOut, 32)

		product, _ := u.DataService.GeProduct(receipt.ProductName)
		if err != nil {
			u.Log.Errorf("Ошибка получения продукта: %s %v", receipt.ProductName, err.Error())
		}
		pointName := strings.Split(receipt.WarehouseName, ",")
		nfd, err := strconv.Atoi(receipt.NumberFiscalDocument)
		if err != nil {
			u.Log.Errorf("Ошибка привидения номера документа к int: %s %v", receipt.ProductName, err.Error())
		}

		if checkReceiptErr == nil && document.Link != "" {
			dateTime, err := time.Parse("2006-01-02T15:04:05", receipt.Date)
			if err != nil {
				u.Log.Errorf("Ошибка разбора даты: %s %v", receipt.ProductName, err.Error())
			}
			dateR := time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), dateTime.Hour(), dateTime.Minute(), dateTime.Second(), 11, time.Local)
			rc := store.Receipt{
				DateTime:             dateR.Format(time.RFC3339Nano),
				FiscalDocumentNumber: nfd,
				KktRegId:             receipt.ManufNumberKKT,
				Link:                 document.Link,
				Name:                 receipt.ProductName,
				Ofd:                  document.Ofd,
				Price:                int(priceSellOut * 100),
				PriceSellIn:          int(priceSellIn * 100),
				ProductId:            product.ID,
				Quantity:             receipt.Quantity,
				CreatedAt:            time.Now().Format(time.RFC3339Nano),
				UpdatedAt:            time.Now().Format(time.RFC3339Nano),
				PointName:            pointName[0],
				SupplerName:          receipt.InnProvider,
				TotalSum:             document.TotalSum,
			}
			if product.ID != "" && product.Export {
				receipts = append(receipts, rc)
			} else {
				_, err := u.DataService.CreateReceipt(rc)
				if err != nil {
					u.Log.Errorf("Ошибка сохранения чека: %v", receipt.ProductName, err)
				}
			}
		} else {
			rc := store.ReceiptN{
				DatePay:      receipt.Date,
				Manufacture:  receipt.ManufacturerName,
				Name:         receipt.ProductName,
				Number:       receipt.DocumentNumber,
				PointName:    pointName[0],
				PriceSellIn:  int(priceSellIn * 100),
				PriceSellOut: int(priceSellOut * 100),
				ProductId:    product.ID,
				Quantity:     receipt.Quantity,
				SupplerName:  receipt.ProviderName,
			}
			if product.ID != "" && product.Export {
				receiptsN = append(receiptsN, rc)
			} else {
				_, err := u.DataService.CreateReceiptN(rc)
				if err != nil {
					u.Log.Errorf("Ошибка сохранения чека: %v", receipt.ProductName, err)
				}
			}
		}
	}

	if len(receipts) > 0 {
		u.MP.SendReceipt(receipts)
	}
	if len(receiptsN) > 0 {
		u.MP.SendReceiptN(receiptsN)
	}
}

func (u *uniFarm) checkReceipt(productName string, fd string, datePay time.Time, totalPrice int) (document model.Document, err error) {
	productName = u.cut(strings.ToLower(productName), 32)
	if receipts, ok := u.cache.Get(productName); ok {
		switch t := receipts.(type) {
		case []platformOfd.Receipt:
			for _, v := range t {
				if fd == v.FD {
					document.Link = v.Link
					document.TotalSum = v.Price
					document.Ofd = "platformofd"
				}
			}
		default:
			u.Log.Errorf("Обработка данной ОФД не доступна")
		}
	}
	return document, err
}

func (u *uniFarm) getMachProduct() {
	matchProducts := u.MP.GetMatchProduct()
	for _, matchProduct := range matchProducts.Data {
		product := store.Product{
			ID:          matchProduct.ID,
			Name:        matchProduct.Name,
			Export:      matchProduct.Export,
			Manufacture: matchProduct.Manufacturer,
		}
		u.DataService.CreateProduct(product)
	}
}

func (u *uniFarm) cut(text string, limit int) string {
	runes := []rune(text)
	if len(runes) >= limit {
		return string(runes[:limit])
	}
	return text
}
