package format

import (
	"Pixel/config"
	"Pixel/core/model"
	"Pixel/store"
	"Pixel/store/service"
	"database/sql"
	"fmt"
	"github.com/PharmaSpace/OfdYa"
	"github.com/PharmaSpace/ofdru"
	"github.com/PharmaSpace/oneofd"
	"github.com/PharmaSpace/platformOfd"
	"github.com/PharmaSpace/sbis"
	"github.com/PharmaSpace/taxcom"
	_ "github.com/denisenkom/go-mssqldb"
	serviceLib "github.com/kardianos/service"
	_ "github.com/mattn/go-adodb"
	"github.com/patrickmn/go-cache"
	"strconv"
	"strings"
	"time"
)

type unico struct {
	date        time.Time
	Config      *config.Config
	DataService *service.DataStore
	MP          *service.Marketpalce
	Log         serviceLib.Logger
	cache       *cache.Cache
}

type UnicoProduct struct {
	PharmacyID      string
	PharmacyAddress string
	Name            string
	Supplier        string
	SupplierINN     string
	Manufacturer    string
	CountryOfOrigin string
	ShipmentNumber  string
	Series          string
	Inventory       float64
	EAN             string
}

type UnicoOrder struct {
	PharmacyID      string
	PharmacyAddress string
	Date            string
	KKM             string
	InvoiceNumber   string
	Manufacturer    string
	Supplier        string
	SupplierINN     string
	Name            string
	PriceWoVat      float64
	PriceWVat       float64
	Vat             float64
	TotalPrice      float64
	TotalNumber     string
	ShipmentNumber  string
	Series          string
}

func Unico(c *config.Config, dataService *service.DataStore, mp *service.Marketpalce, cache *cache.Cache, log serviceLib.Logger, date time.Time) *unico {
	return &unico{Config: c, DataService: dataService, cache: cache, MP: mp, Log: log, date: date}
}

func (u *unico) connect() (*sql.DB, error) {
	/*query := url.Values{}
	query.Add("app name", "Pixel")

	usql := &url.URL{
		Scheme: "sqlserver",
		User:   url.UserPassword(u.Config.UnicoOptions.Username, u.Config.UnicoOptions.Password),
		Host:   fmt.Sprintf("%s:%d", u.Config.UnicoOptions.Host, 1433),
		// Path:  instance, // if connecting to an instance instead of a port
		RawQuery: query.Encode(),
	}*/
	var (
		conn *sql.DB
		err  error
	)

	if u.Config.UnicoOptions.SqlDriver == "sqlserver" {
		conn, err = sql.Open("sqlserver", u.Config.UnicoOptions.ConnString)
	} else {
		conn, err = sql.Open("adodb", u.Config.UnicoOptions.ConnString)
	}
	return conn, err
}

func (u *unico) Parse() {
	u.sendProduct()
	u.getMachProduct()

	receiptsN := []store.ReceiptN{}
	receipts := []store.Receipt{}

	receiptsRow := u.getReceipt()
	for _, receipt := range receiptsRow {
		datePay, _ := time.Parse("02.01.2006", receipt.Date)
		document, checkReceiptErr := u.checkReceipt(receipt.Name, receipt.InvoiceNumber, datePay, int(receipt.TotalPrice*100))
		product, err := u.DataService.GeProduct(receipt.Name)
		if err != nil {
			u.Log.Errorf("Ошибка получения продукта: %s %v", receipt.Name, err.Error())
		}
		pointName := strings.Split(receipt.PharmacyID, ",")
		nfd, err := strconv.Atoi(receipt.InvoiceNumber)
		if err != nil {
			u.Log.Errorf("Ошибка привидения номера документа к int: %s %v", receipt.Name, err.Error())
		}

		if checkReceiptErr == nil && document.Link != "" {
			dateTime, err := time.Parse("02.01.2006", receipt.Date)
			if err != nil {
				u.Log.Errorf("Ошибка разбора даты: %s %v", receipt.Name, err.Error())
			}
			dateR := time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), dateTime.Hour(), dateTime.Minute(), dateTime.Second(), 11, time.Local)
			rc := store.Receipt{
				DateTime:             dateR.Format(time.RFC3339Nano),
				FiscalDocumentNumber: nfd,
				KktRegId:             document.KktRegId,
				Link:                 document.Link,
				Name:                 receipt.Name,
				Ofd:                  document.Ofd,
				Price:                int(receipt.PriceWVat * 100),
				PriceSellIn:          int(receipt.PriceWVat * 100),
				ProductId:            product.ID,
				Quantity:             receipt.TotalNumber,
				CreatedAt:            time.Now().Format(time.RFC3339Nano),
				UpdatedAt:            time.Now().Format(time.RFC3339Nano),
				PointName:            pointName[0],
				SupplerName:          receipt.Supplier,
				TotalSum:             document.TotalSum,
			}
			if product.ID != "" && product.Export {
				receipts = append(receipts, rc)
			}
		} else {
			rc := store.ReceiptN{
				DatePay:      receipt.Date,
				Manufacture:  receipt.Manufacturer,
				Name:         receipt.Name,
				Number:       receipt.InvoiceNumber,
				PointName:    pointName[0],
				PriceSellIn:  0,
				PriceSellOut: int(receipt.PriceWVat * 100),
				ProductId:    product.ID,
				Quantity:     receipt.TotalNumber,
				SupplerName:  receipt.Supplier,
			}
			if product.ID != "" && product.Export {
				receiptsN = append(receiptsN, rc)
			}
		}
	}
}

func (u *unico) sendProduct() {
	conn, err := u.connect()
	if err != nil {
		u.Log.Errorf("Ошибка подключения к MSSQL %v", err)
	}
	defer conn.Close()

	query := fmt.Sprintf("select  T.Name AS 'НАИМЕНОВАНИЕ ТОВАРА', F.NameFactory AS 'ПРОИЗВОДИТЕЛЬ',CAST(L.Quantity as decimal(12, 2)) AS 'КОЛ-ВО', '' AS 'НОМЕР ПАРТИИ',L.Serial AS 'СЕРИЙНЫЙ НОМЕР', PD.Podr AS 'ID АПТЕКИ', C.ShortName AS 'ПОСТАВЩИК' from PDoc PD INNER JOIN ListDoc L on L.CodDoc = PD.Cod INNER JOIN TMC T on T.Cod = L.CodTMC INNER JOIN Factory F on F.Cod = T.CodFactory INNER JOIN Country CT on CT.Cod = F.CodCountry INNER JOIN v_Division DIV on DIV.Cod = PD.Podr LEFT JOIN ListDoc Lco on Lco.Cod = L.CodOrig LEFT JOIN PDoc PDco on PDco.Cod = Lco.CodDoc INNER JOIN Client C on C.Cod = PDco.Client WHERE PD.DNacl >= dbo.Fn_UNI_ConvertDateFromSQL('%s') AND PD.DNacl <= dbo.Fn_UNI_ConvertDateFromSQL('%s')  AND PD.DebKred = 2  AND PD.IsCassa > 0 GROUP BY PD.Podr, DIV.ShortName, T.Name, C.ShortName, C.INN, F.NameFactory, CT.Name, L.Serial, L.Quantity, T.ScanCod ORDER BY T.Name ASC", u.date.Format("02.01.2006"), u.date.Format("02.01.2006"))
	rows, err := conn.Query(query)
	if err != nil {
		u.Log.Errorf("Ошибка запроса в получении товаров %v", err)
	}
	defer rows.Close()

	products := make([]service.Product, 0)
	for rows.Next() {
		product := new(service.Product)
		err := rows.Scan(&product.Name, &product.Manufacturer, &product.Stock, &product.PartNumber, &product.Serial, &product.WarehouseName, &product.SupplerName)
		if err != nil {
			u.Log.Errorf("Ошибка в переменной product %v", err)
		}
		products = append(products, *product)
	}

	u.MP.SendProduct(products)
}

func (u *unico) getReceipt() (receipts []*UnicoOrder) {
	conn, err := u.connect()
	if err != nil {
		u.Log.Errorf("Ошибка подключения к MSSQL %v", err)
	}
	defer conn.Close()

	query := fmt.Sprintf("select PD.Podr AS 'ID АПТЕКИ', DIV.ShortName AS 'АПТЕКА', FORMAT( CONVERT(datetime, DC.Date - 36163), 'dd.MM.yyyy', 'ru-RU' ) AS 'ДАТА ПРОДАЖИ', DCI.FPDKKM AS 'НОМЕР ККМ', DC.CodCotter AS 'НОМЕР ЧЕКА', F.NameFactory AS 'ПРОИЗВОДИТЕЛЬ', C.ShortName AS 'ПОСТАВЩИК', C.INN AS 'ИНН ПОСТ.', DC.NameTMC AS 'НАИМЕНОВАНИЕ ТОВАРА', L.PriceSaleNaked AS 'ЦЕНА ПРОДАЖИ БЕЗ НДС', L.PriceSale AS 'ЦЕНА ПРОДАЖИ С НДС', L.SumNDSSale AS 'СУММА НДС НА ЕДИНИЦУ', L.SumSale AS 'ОБЩАЯ СУММА ПО ЧЕКУ', CAST(L.Quantity as decimal(12, 2)) AS 'КОЛ-ВО', '' AS 'НОМЕР ПАРТИИ', L.Serial AS 'СЕРИЙНЫЙ НОМЕР' from PDoc PD INNER JOIN ListDoc L on L.CodDoc = PD.Cod INNER JOIN DupCott DC on DC.CodListDocKredit = L.Cod AND DC.CodPodr = PD.Podr AND DC.IsStorno = 0 and DC.IsErrCot = 0 and DC.IsDelete = 0 LEFT JOIN DupCotInfo DCI on DCI.CodGuid = DC.DCIGuid AND DCI.Date = DC.Date AND DCI.NumCash = DC.NumCash INNER JOIN TMC T on T.Cod = DC.CodTMC INNER JOIN Factory F on F.Cod = T.CodFactory INNER JOIN Country CT on CT.Cod = F.CodCountry INNER JOIN v_Division DIV on DIV.Cod = PD.Podr LEFT JOIN ListDoc Lco on Lco.Cod = L.CodOrig LEFT JOIN PDoc PDco on PDco.Cod = Lco.CodDoc INNER JOIN Client C on C.Cod = PDco.Client WHERE PD.DNacl >= dbo.Fn_UNI_ConvertDateFromSQL('%s') AND PD.DNacl <= dbo.Fn_UNI_ConvertDateFromSQL('%s') AND PD.DebKred = 2 AND PD.IsCassa > 0  GROUP BY PD.Podr, DIV.ShortName, DC.Date, DCI.FPDKKM, DC.CodCotter, F.NameFactory, C.ShortName, C.INN, DC.NameTMC, L.PriceSaleNaked, L.PriceSale, L.SumNDSSale, L.SumSale, L.Quantity, L.Serial ORDER BY DC.NameTMC ASC", u.date.Format("02.01.2006"), u.date.Format("02.01.2006"))
	rows, err := conn.Query(query)
	if err != nil {
		u.Log.Errorf("Ошибка запроса в получении заказов %v", err)
	}
	defer rows.Close()

	orders := make([]*UnicoOrder, 0)
	for rows.Next() {
		order := new(UnicoOrder)
		err := rows.Scan(&order.PharmacyID, &order.PharmacyAddress, &order.Date, &order.KKM, &order.InvoiceNumber, &order.Manufacturer, &order.Supplier, &order.SupplierINN, &order.Name, &order.PriceWoVat, &order.PriceWVat, &order.Vat, &order.TotalPrice, &order.TotalNumber, &order.ShipmentNumber, &order.Series)
		if err != nil {
			u.Log.Errorf("Ошибка в переменной order %v", err)
		}
		orders = append(orders, order)
	}

	return orders

}

func (u *unico) checkReceipt(productName string, fd string, datePay time.Time, totalPrice int) (document model.Document, err error) {
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
		case []oneofd.Receipt:
			for _, v := range t {
				if fd == v.FD || fd == v.FP {
					document.Link = v.Link
					document.TotalSum = v.Price
					document.Ofd = "1ofd"
				}
			}
		case []ofdru.Receipt:
			for _, v := range t {
				if fd == v.FD || fd == v.FP {
					document.Link = v.Link
					document.TotalSum = v.Price
					document.Ofd = "ofdru"
				}
			}
		case []OfdYa.Receipt:
			for _, v := range t {
				if fd == v.FD || fd == v.FP {
					document.Link = v.Link
					document.TotalSum = v.Price
					document.Ofd = "ofd-ya"
				}
			}
		case []*sbis.Receipt:
			for _, v := range t {
				if fd == strconv.Itoa(v.RequestNumber) || totalPrice == v.TotalSum {
					for _, i := range v.Items {
						if i.Name == strings.ToUpper(productName) {
							date, _ := time.Parse("2006-01-02T15:04:05", v.ReceiveDateTime)
							document.DateTime = date.Unix()
							document.Link = v.Url
							document.FiscalDocumentNumber = v.FiscalDocumentNumber
							document.KktRegId = v.KktRegID
							document.ProductPrice = i.Price
							document.TotalSum = v.TotalSum
							document.Ofd = "sbis"
						}
					}
				}
			}
		case []taxcom.Receipt:
			for _, v := range t {
				if fd == v.FD || fd == v.FP || fd == strconv.Itoa(v.DocumentNumber) {
					fiscalDocumentNumber, _ := strconv.Atoi(v.FD)
					document.KktRegId = v.KktRegId
					document.Link = v.Link
					document.TotalSum = v.Price
					document.FiscalDocumentNumber = fiscalDocumentNumber
					document.Ofd = "taxcom"
				}
			}
		default:
			u.Log.Errorf("Обработка данной ОФД не доступна")
		}
	}
	return document, err
}

func (u *unico) cut(text string, limit int) string {
	runes := []rune(text)
	if len(runes) >= limit {
		return string(runes[:limit])
	}
	return text
}

func (u *unico) getMachProduct() {
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
