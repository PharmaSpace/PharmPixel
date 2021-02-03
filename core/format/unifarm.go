package format

import (
	serviceLib "github.com/kardianos/service"
	"github.com/patrickmn/go-cache"
	"log"
	"pixel/config"
	"pixel/core/model"
	"pixel/sentry"
	"pixel/store/service"
	"strconv"
	"strings"
	"time"
)

// UniFarm структура
type UniFarm struct {
	sentry           *sentry.Sentry
	date             time.Time
	Config           *config.Config
	matchingOfdCache *cache.Cache
	matchingErpCache *cache.Cache
	MP               service.MarketPlaceInterface
	Log              serviceLib.Logger
	cache            *cache.Cache
}

// NewUniFarm интеграциия
func NewUniFarm(c *config.Config, mp service.MarketPlaceInterface, ca *cache.Cache, l serviceLib.Logger, date time.Time, s *sentry.Sentry) *UniFarm {
	matchingOfdCache := cache.New(5*time.Minute, 10*time.Minute)
	matchingErpCache := cache.New(5*time.Minute, 10*time.Minute)

	return &UniFarm{Config: c, MP: mp, Log: l, cache: ca, matchingOfdCache: matchingOfdCache, matchingErpCache: matchingErpCache, date: date, sentry: s}
}

// GetCache получение кеша по ключу
func (u *UniFarm) GetCache(key string) (interface{}, bool) {
	if u.cache == nil {
		return nil, false
	}
	return u.cache.Get(key)
}

// GetOFDCache получение кеша по ключу
func (u *UniFarm) GetOFDCache(key string) (interface{}, bool) {
	if u.matchingOfdCache == nil {
		return nil, false
	}
	return u.matchingOfdCache.Get(key)
}

// GetERPCache получение кеша по ключу
func (u *UniFarm) GetERPCache(key string) (interface{}, bool) {
	if u.matchingErpCache == nil {
		return nil, false
	}
	return u.matchingErpCache.Get(key)
}

// SetOFDCache получение кеша по ключу
func (u *UniFarm) SetOFDCache(key string, val interface{}, duration time.Duration) {
	if u.matchingOfdCache == nil {
		return
	}
	u.matchingOfdCache.Set(key, val, duration)
}

// SetERPCache получение кеша по ключу
func (u *UniFarm) SetERPCache(key string, val interface{}, duration time.Duration) {
	if u.matchingErpCache == nil {
		return
	}
	u.matchingOfdCache.Set(key, val, duration)
}

// GetMP иницилизация МП
func (u *UniFarm) GetMP() service.MarketPlaceInterface {
	return u.MP
}

// GetDate получение даты
func (u *UniFarm) GetDate() time.Time {
	return u.date
}

// Parse разбор данных
func (u *UniFarm) Parse() {
	uni := service.NewUniFarm(u.Config.UniFarmOptions.Username, u.Config.UniFarmOptions.Password)
	var err error
	err = getMatchProducts(u)
	if err != nil {
		u.sentry.Error(err)
	}
	products := u.convertProducts(uni.GetProduct(u.date), u.date)

	// данные из кеша по чекам из ОФД
	ofdRecieptCacheItems := u.cache.Items()
	// чеки из ЕРП
	erpReceipts := u.convertReceipts(uni.GetReceipt(u.date))

	receipts, checkOfdProductNames := nameMatching(u, erpReceipts, ofdRecieptCacheItems)

	if len(receipts) > 0 {
		err = u.MP.SendReceipt(receipts)
		if err != nil {
			u.sentry.Error(err)
			log.Printf("[ERROR] Ошибка отправки чеков %v", err)
		}
	}

	if len(checkOfdProductNames) > 0 {
		// отправляем товары из ofd на матчинг
		productsForMatching := OfdProductsForMatching(checkOfdProductNames)
		err = u.MP.SendOfdProducts(productsForMatching, true, false)
		if err != nil {
			u.sentry.Error(err)
			log.Printf("[ERROR] Ошибка отправки продуктов %v", err)
		}
	}

	// отправляем товары из erp на матчинг
	productsForMatching := ErpProductsForMatching(u, products)
	err = u.MP.SendOfdProducts(productsForMatching, false, true)
	if err != nil {
		u.sentry.Error(err)
		log.Printf("[ERROR] Ошибка отправки продуктов %v", err)
	}
}

func (u *UniFarm) convertReceipts(receipts []service.UniFarmReceipt) []*model.Receipt {
	orders := make([]*model.Receipt, 0, len(receipts))

	for _, receipt := range receipts {
		pointName := strings.Split(receipt.WarehouseName, ",")
		// TODO: Проверить правильность конвертации и посмотреть можно ли заполнить пустые поля
		orders = append(orders, &model.Receipt{
			PharmacyID:      pointName[0],
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
			Series:          receipt.Serial,
		})
	}

	return orders
}

func (u *UniFarm) convertProducts(uniFarmProducts []service.UniFarmProduct, t time.Time) []service.Product {
	products := make([]service.Product, 0)
	for _, product := range uniFarmProducts {
		stock, _ := strconv.ParseFloat(product.Stock, 32)
		pointName := strings.Split(product.WarehouseName, ",")

		products = append(products, service.Product{
			Name:          product.ProductName,
			Manufacturer:  product.ManufacturerName,
			WarehouseName: pointName[0],
			SupplerName:   product.SupplierName,
			SupplerInn:    product.SupplierINN,
			Stock:         stock,
			PartNumber:    product.PartNumber,
			Serial:        product.Serial,
			CreatedAt:     t.Format(time.RFC3339),
		})
	}
	return products
}
