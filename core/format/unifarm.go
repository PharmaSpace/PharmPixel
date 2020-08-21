package format

import (
	"Pixel/config"
	"Pixel/core/model"
	"Pixel/store/service"
	serviceLib "github.com/kardianos/service"
	"github.com/patrickmn/go-cache"
	"strings"
	"time"
)

type uniFarm struct {
	date        time.Time
	Config      *config.Config
	matchingOfdCache *cache.Cache
	matchingErpCache *cache.Cache
	MP          *service.Marketpalce
	Log         serviceLib.Logger
	cache       *cache.Cache
}

func UniFarm(c *config.Config, dataService *service.DataStore, mp *service.Marketpalce, ca *cache.Cache, log serviceLib.Logger, date time.Time) *uniFarm {
	matchingOfdCache := cache.New(5*time.Minute, 10*time.Minute)
	matchingErpCache := cache.New(5*time.Minute, 10*time.Minute)

	return &uniFarm{Config: c, MP: mp, Log: log, cache: ca, matchingOfdCache: matchingOfdCache, matchingErpCache: matchingErpCache, date: date}
}


func (u *uniFarm) GetCache(key string) (interface{}, bool) {
	if u.cache == nil {
		return nil, false
	}
	return u.cache.Get(key)
}

func (u *uniFarm) GetOFDCache(key string) (interface{}, bool) {
	if u.matchingOfdCache == nil {
		return nil, false
	}
	return u.matchingOfdCache.Get(key)
}

func (u *uniFarm) GetERPCache(key string) (interface{}, bool) {
	if u.matchingErpCache == nil {
		return nil, false
	}
	return u.matchingErpCache.Get(key)
}

func (u *uniFarm) SetOFDCache(key string, val interface{}, duration time.Duration) {
	if u.matchingOfdCache == nil {
		return
	}
	u.matchingOfdCache.Set(key, val, duration)
}


func (u *uniFarm) SetERPCache(key string, val interface{}, duration time.Duration) {
	if u.matchingErpCache == nil {
		return
	}
	u.matchingOfdCache.Set(key, val, duration)
}

func (u *uniFarm) GetMP() service.MarketPlaceInterface {
	return u.MP
}

func (u *uniFarm) GetDate() time.Time {
	return u.date
}

func (u *uniFarm) Parse() {
	uni := service.UniFarm(u.Config.UniFarmOptions.Username, u.Config.UniFarmOptions.Password)
	getMatchProducts(u)

	products := u.convertProducts(uni.GetProduct(u.date))

	// данные из кеша по чекам из ОФД
	ofdRecieptCacheItems := u.cache.Items()
	// чеки из ЕРП
	erpReceipts := u.convertReceipts(uni.GetReceipt(u.date))

	receipts, checkOfdProductNames := nameMatching(u, erpReceipts, ofdRecieptCacheItems)

	if len(receipts) > 0 {
		u.MP.SendReceipt(receipts)
	}

	if len(checkOfdProductNames) > 0 {
		// отправляем товары из ofd на матчинг
		productsForMatching := OfdProductsForMatching(checkOfdProductNames)
		u.MP.SendOfdProducts(productsForMatching, true, false)
	}

	// отправляем товары из erp на матчинг
	productsForMatching := ErpProductsForMatching(u, products)
	u.MP.SendOfdProducts(productsForMatching, false, true)
}

func (u *uniFarm)  convertReceipts(receipts  []service.UniFarmReceipt) []*model.Receipt {
	orders := make([]*model.Receipt, 0)

	for _, receipt := range receipts {
		pointName := strings.Split(receipt.WarehouseName, ",")
		// TODO: Проверить правильность конвертации и посмотреть можно ли заполнить пустые поля
		orders = append(orders, &model.Receipt{
			PharmacyID:     pointName[0],
			PharmacyAddress: pointName[1],
			Date:            receipt.Date,
			KKM:             receipt.NumberKKT,
			InvoiceNumber:   receipt.NumberFiscalDocument,
			Manufacturer:    receipt.ManufacturerName,
			Supplier:        receipt.ProviderName,
			SupplierINN:     receipt.InnProvider,
			Name:            receipt.ProductName,
			PriceWoVat:      receipt.PriceSellIn,
			PriceWVat:       receipt.PriceSellOut,
			Vat:             "",
			TotalPrice:      receipt.SumPriceSellOut,
			TotalNumber:     "",
			ShipmentNumber:  receipt.PartNumber,
			Series:         receipt.Serial,
		})
	}

	return orders
}

func (u *uniFarm) convertProducts(uniFarmProducts []service.UniFarmProduct) []service.Product {
	products := make([]service.Product, 0)
	for _, product := range uniFarmProducts {
		products = append(products, service.Product{
			Name:         product.ProductName,
			Manufacturer: product.ManufacturerName,
			Stock:        0,
			PartNumber:   product.PartNumber,
			Serial:       product.Serial,
		})
	}
	return products
}
