package format

import (
	"database/sql"
	"fmt"

	// Register some standard stuff
	_ "github.com/denisenkom/go-mssqldb"
	serviceLib "github.com/kardianos/service"
	// Register some standard stuff
	_ "github.com/mattn/go-adodb"
	"github.com/patrickmn/go-cache"
	"log"
	"pixel/config"
	"pixel/core/model"
	"pixel/db"
	"pixel/store/service"
	"strings"
	"time"
)

// Unico структура для интерации unico
type Unico struct {
	date             time.Time
	Config           *config.Config
	MP               service.MarketPlaceInterface
	Log              serviceLib.Logger
	cache            *cache.Cache
	matchingOfdCache *cache.Cache
	matchingErpCache *cache.Cache
	DB               db.DB
	Errs             []error
}

// NewUnico формат подключение аптеки
func NewUnico(c *config.Config, mp service.MarketPlaceInterface, d db.DB, rCache *cache.Cache, l serviceLib.Logger, date time.Time) *Unico {
	matchingOfdCache := cache.New(5*time.Minute, 10*time.Minute)
	matchingErpCache := cache.New(5*time.Minute, 10*time.Minute)

	return &Unico{Config: c, cache: rCache, MP: mp, DB: d, Log: l, date: date, matchingOfdCache: matchingOfdCache, matchingErpCache: matchingErpCache}
}

// GetCache получение кеша по ключу
func (u *Unico) GetCache(key string) (interface{}, bool) {
	if u.cache == nil {
		return nil, false
	}
	return u.cache.Get(key)
}

// GetOFDCache получение кеша по ключу
func (u *Unico) GetOFDCache(key string) (interface{}, bool) {
	if u.matchingOfdCache == nil {
		return nil, false
	}
	return u.matchingOfdCache.Get(key)
}

// GetERPCache получение кеша по ключу
func (u *Unico) GetERPCache(key string) (interface{}, bool) {
	if u.matchingErpCache == nil {
		return nil, false
	}
	return u.matchingErpCache.Get(key)
}

// SetOFDCache получение кеша по ключу
func (u *Unico) SetOFDCache(key string, val interface{}, duration time.Duration) {
	if u.matchingOfdCache == nil {
		return
	}
	u.matchingOfdCache.Set(key, val, duration)
}

// SetERPCache получение кеша по ключу
func (u *Unico) SetERPCache(key string, val interface{}, duration time.Duration) {
	if u.matchingErpCache == nil {
		return
	}
	u.matchingOfdCache.Set(key, val, duration)
}

// GetMP иницилизация МП
func (u *Unico) GetMP() service.MarketPlaceInterface {
	return u.MP
}

// GetDate получение даты
func (u *Unico) GetDate() time.Time {
	return u.date
}

// Parse парсим данные
func (u *Unico) Parse() {
	u.matchingErpCache.Flush()
	u.matchingOfdCache.Flush()
	var err error
	getMatchProducts(u)

	// данные из кеша по чекам из ОФД
	ofdRecieptCacheItems := u.cache.Items()
	// чеки из ЕРП
	erpReceipts := u.getReceipts()
	receipts, checkOfdProductNames := nameMatching(u, erpReceipts, ofdRecieptCacheItems)
	if len(receipts) > 0 {
		err = u.MP.SendReceipt(receipts)
		if err != nil {
			log.Printf("[ERROR] отпрвка чеков %v", err)
		}
	}

	if len(checkOfdProductNames) > 0 {
		// отправляем товары из ofd на матчинг
		productsForMatching := OfdProductsForMatching(checkOfdProductNames)
		err = u.MP.SendOfdProducts(productsForMatching, true, false)
		if err != nil {
			log.Printf("[ERROR] отпрвка продуктов  %v", err)
		}
	}
	// отправляем товары из erp на матчинг
	productsForMatching := ErpProductsForMatching(u, u.getErpProducts())
	err = u.MP.SendOfdProducts(productsForMatching, false, true)
	if err != nil {
		log.Printf("[ERROR] отпрвка продуктов %v", err)
	}
}

// OfdProductsForMatching подготовка продуктов к отправке
func (u *Unico) OfdProductsForMatching(ofdProductNames []string) {
	products := make([]service.Product, 0, len(ofdProductNames))
	for _, name := range ofdProductNames {
		products = append(products, service.Product{
			Name: strings.ToLower(name),
		})
	}
	err := u.MP.SendOfdProducts(products, true, false)
	if err != nil {
		log.Printf("[ERROR] отправка продуктов %v", err)
	}
}

func (u *Unico) getErpProducts() []service.Product {
	rows, err := u.DB.Query(fmt.Sprintf(`select  T.Name AS 'НАИМЕНОВАНИЕ ТОВАРА', F.NameFactory AS 'ПРОИЗВОДИТЕЛЬ',L.Quantity AS 'КОЛ-ВО', '' AS 'НОМЕР ПАРТИИ',L.Serial AS 'СЕРИЙНЫЙ НОМЕР', PD.Podr AS 'ID АПТЕКИ', C.ShortName AS 'ПОСТАВЩИК', C.INN AS 'ИНН ПОСТ.'
	from PDoc PD INNER JOIN ListDoc L on L.CodDoc = PD.Cod INNER JOIN TMC T on T.Cod = L.CodTMC 
	INNER JOIN Factory F on F.Cod = T.CodFactory 
	INNER JOIN Country CT on CT.Cod = F.CodCountry 
	INNER JOIN v_Division DIV on DIV.Cod = PD.Podr 
	LEFT JOIN ListDoc Lco on Lco.Cod = L.CodOrig 
	LEFT JOIN PDoc PDco on PDco.Cod = Lco.CodDoc 
	INNER JOIN Client C on C.Cod = PDco.Client 
	WHERE PD.DNacl >= dbo.Fn_UNI_ConvertDateFromSQL('%s') 
	AND PD.DNacl <= dbo.Fn_UNI_ConvertDateFromSQL('%s')  
	AND PD.DebKred = 2  AND PD.IsCassa > 0 GROUP BY PD.Podr, DIV.ShortName, T.Name, C.ShortName, C.INN, F.NameFactory, CT.Name, L.Serial, L.Quantity, T.ScanCod 
	ORDER BY T.Name ASC`, u.date.Format("02.01.2006"), u.date.Format("02.01.2006")))
	if err != nil {
		err = u.Log.Errorf("Ошибка запроса в получении товаров %v", err)
		if err != nil {
			log.Printf("Ошибка запроса в получении товаров %v", err)
		}
		u.Errs = append(u.Errs, err)
	}
	if rows == nil {
		return nil
	}
	defer rows.Close()

	products := make([]service.Product, 0)
	for rows.Next() {
		product := service.Product{}
		err := rows.Scan(&product.Name, &product.Manufacturer, &product.Stock, &product.PartNumber, &product.Serial, &product.WarehouseName, &product.SupplerName, &product.SupplerInn)
		if err != nil {
			err = u.Log.Errorf("Ошибка в переменной product %v", err)
			if err != nil {
				log.Printf("Ошибка в переменной product %v", err)
			}
		}
		product.CreatedAt = u.date.Format(time.RFC3339Nano)
		products = append(products, product)
	}

	return products
}

func (u *Unico) getReceipts() (receipts []*model.Receipt) {
	var err error

	rows, err := u.DB.Query(fmt.Sprintf(`select PD.Podr AS 'ID АПТЕКИ', DIV.ShortName AS 'АПТЕКА',
	FORMAT(DC.Date, 'dd/MM/yyyy') AS 'ДАТА ПРОДАЖИ',
	DCI.FPDKKM AS 'НОМЕР ККМ', DC.CodCotter AS 'НОМЕР ЧЕКА',
	F.NameFactory AS 'ПРОИЗВОДИТЕЛЬ', C.ShortName AS 'ПОСТАВЩИК',
	C.INN AS 'ИНН ПОСТ.', DC.NameTMC AS 'НАИМЕНОВАНИЕ ТОВАРА',
	L.PriceSaleNaked AS 'ЦЕНА ПРОДАЖИ БЕЗ НДС',
	L.PriceSale AS 'ЦЕНА ПРОДАЖИ С НДС', L.SumNDSSale AS 'СУММА НДС НА ЕДИНИЦУ',
	L.SumSale AS 'ОБЩАЯ СУММА ПО ЧЕКУ', CAST(L.Quantity as decimal(12, 2)) AS 'КОЛ-ВО',
	'' AS 'НОМЕР ПАРТИИ', L.Serial AS 'СЕРИЙНЫЙ НОМЕР' 
	from PDoc PD INNER JOIN ListDoc L on L.CodDoc = PD.Cod 
	INNER JOIN DupCott DC on DC.CodListDocKredit = L.Cod AND DC.CodPodr = PD.Podr AND DC.IsStorno = 0 and DC.IsErrCot = 0 and DC.IsDelete = 0 
	LEFT JOIN DupCotInfo DCI on DCI.CodGuid = DC.DCIGuid AND DCI.Date = DC.Date AND DCI.NumCash = DC.NumCash 
	INNER JOIN TMC T on T.Cod = DC.CodTMC INNER JOIN Factory F on F.Cod = T.CodFactory 
	INNER JOIN Country CT on CT.Cod = F.CodCountry INNER JOIN v_Division DIV on DIV.Cod = PD.Podr 
	LEFT JOIN ListDoc Lco on Lco.Cod = L.CodOrig LEFT JOIN PDoc PDco on PDco.Cod = Lco.CodDoc 
	INNER JOIN Client C on C.Cod = PDco.Client WHERE PD.DNacl >= dbo.Fn_UNI_ConvertDateFromSQL('%s') 
	AND PD.DNacl <= dbo.Fn_UNI_ConvertDateFromSQL('%s') 
	AND PD.DebKred = 2 AND PD.IsCassa > 0  
	GROUP BY PD.Podr, DIV.ShortName, DC.Date, DCI.FPDKKM, DC.CodCotter, F.NameFactory, C.ShortName, C.INN, DC.NameTMC, L.PriceSaleNaked, L.PriceSale, L.SumNDSSale, L.SumSale, L.Quantity, L.Serial 
	ORDER BY DC.NameTMC ASC`, u.date.Format("02.01.2006"), u.date.Format("02.01.2006")))
	if err != nil {
		err = u.Log.Errorf("Ошибка запроса в получении заказов %v", err)
		if err != nil {
			log.Printf("Ошибка запроса в получении заказов %v", err)
		}
		u.Errs = append(u.Errs, err)
	}
	if rows == nil {
		return
	}
	defer rows.Close()

	orders := make([]*model.Receipt, 0)
	for rows.Next() {
		order := new(model.Receipt)
		date := ""
		err := rows.Scan(&order.PharmacyID, &order.PharmacyAddress, &date, &order.KKM, &order.InvoiceNumber, &order.Manufacturer, &order.Supplier, &order.SupplierINN, &order.Name, &order.PriceWoVat, &order.PriceWVat, &order.Vat, &order.TotalPrice, &order.TotalNumber, &order.ShipmentNumber, &order.Series)
		if err != nil {
			err = u.Log.Errorf("Ошибка в переменной order %v", err)
			if err != nil {
				log.Printf("Ошибка в переменной order %v", err)
			}
		}
		orders = append(orders, order)
	}

	return orders

}

// ConnectToErpDB подключение к MSSQL
func ConnectToErpDB(c *config.Config) (db.DB, error) {
	if c.UnicoOptions.SQLDriver == "sqlserver" {
		return sql.Open("sqlserver", c.UnicoOptions.ConnString)
	}
	return sql.Open("adodb", c.UnicoOptions.ConnString)
}
