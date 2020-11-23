package format

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"pixel/config"
	"testing"
)

func TestUnico_getReceipts(t *testing.T) {
	_ = godotenv.Load("../../.env")
	config.Load()

	database, err := ConnectToErpDB(config.Cfg)
	assert.NoError(t, err)

	_, err = database.Query(fmt.Sprintf(`select PD.Podr AS 'ID АПТЕКИ', DIV.ShortName AS 'АПТЕКА',
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
	ORDER BY DC.NameTMC ASC`, "07.10.2020", "07.10.2020"))
	assert.NoError(t, err)
}

/*
import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/joho/godotenv"
	windowsService "github.com/kardianos/service"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"Pixel/config"
	"Pixel/core/format/mock"
	"Pixel/store/service"
	"testing"
	"time"
)

var errCouldntConnectToDB = errors.New("Couldn't connect to db")

var sqlResp = sqlmock.NewRows([]string{
	"pharmascy_id",
	"pharmacy_address",
	"date",
	"kkm",
	"invoice_number",
	"manufacturer",
	"supplier",
	"supplier_inn",
	"name",
	"price_wo_wat",
	"price_w_vat",
	"vat",
	"total_price",
	"total_number",
	"shipment_number",
	"series",
})

func TestUnico_Parse(t *testing.T) {
	_ = godotenv.Overload(".env.test")
	t.Run("err connect to db", func(t *testing.T) {
		logger := windowsService.ConsoleLogger
		config.Load()
		c := cache.New(5*time.Minute, 10*time.Minute)
		setCache(c, ofdItems)
		mpMock := &mock.MarketPlaceMock{}
		dbMock := &mock.DBMock{}
		date := time.Now()

		unico := NewUnico(config.Cfg, mpMock, dbMock, c, logger, date)
		unico.Config.UnicoOptions.SQLDriver = "test"
		mpMock.On("GetMatchProducts").Return(service.MatchProducts{}).Twice()
		dbMock.On("Query", mock2.AnythingOfType("string"), mock2.Anything).
			Return(sql.Rows{}, errCouldntConnectToDB)
		mpMock.On("SendOfdProducts",
			[]service.Product{
				{Name: "ПЕКТУСИН  ТАБ. Х10"},
				{Name: "2/5уп РУМАЛОН Р-Р В/М ВВЕД. АМП."},
			}, true, false).Return().Once()
		mpMock.On("SendOfdProducts", []service.Product(nil), false, true).Once()
		unico.Parse()
		for _, err := range unico.Errs {
			assert.Equal(t, errCouldntConnectToDB, err)
		}
		mpMock.AssertExpectations(t)
	})

	t.Run("match found by invoice num", func(t *testing.T) {
		logger := windowsService.ConsoleLogger
		config.Load()
		c := cache.New(5*time.Minute, 10*time.Minute)
		setCache(c, ofdItems)
		mpMock := &mock.MarketPlaceMock{}
		//dbMock := &mock.DBMock{}
		db, dbMock, _ := sqlmock.New()
		date := time.Now()
		unico := NewUnico(config.Cfg, mpMock, db, c, logger, date)
		unico.Config.UnicoOptions.SQLDriver = "test"
		mpMock.On("GetMatchProducts").
			Return(service.MatchProducts{Data: []service.MatchProductItem{{
				Export:       true,
				ID:           "matchId",
				Name:         "ПЕКТУСИН  ТАБ. Х10",
				Stock:        1,
				SupplierName: "OOO",
				SupplierInn:  "111111111111",
			},
			}}).Once()
		mpMock.On("GetMatchProducts").Return(service.MatchProducts{}).Once()

		dbMock.ExpectQuery("^select*").
			WillReturnRows(sqlResp.AddRow(
				"123",
				"addr",
				"",
				"",
				"92975",
				"man",
				"sup",
				"12222",
				"ПЕКТУСИН  ТАБ. Х10",
				0.0, 0.0, 0.0, 34080.0, "", "", ""))
		mpMock.On("SendReceipt", mock2.AnythingOfType("[]store.Receipt")).Return()
		mpMock.On("SendOfdProducts", []service.Product{{Name: "2/5уп РУМАЛОН Р-Р В/М ВВЕД. АМП."}}, true, false).Return().Once()
		mpMock.On("SendOfdProducts", []service.Product(nil), false, true).Return().Once()
		unico.Parse()
		mpMock.AssertExpectations(t)
	})

	t.Run("match found by total price", func(t *testing.T) {
		logger := windowsService.ConsoleLogger
		config.Load()
		c := cache.New(5*time.Minute, 10*time.Minute)
		setCache(c, ofdItems)
		mpMock := &mock.MarketPlaceMock{}
		//dbMock := &mock.DBMock{}
		db, dbMock, _ := sqlmock.New()
		date := time.Now()
		unico := NewUnico(config.Cfg, mpMock, db, c, logger, date)
		unico.Config.UnicoOptions.SQLDriver = "test"
		mpMock.On("GetMatchProducts").
			Return(service.MatchProducts{Data: []service.MatchProductItem{{
				Export:       true,
				ID:           "matchId",
				Name:         "ПЕКТУСИН  ТАБ. Х10",
				Stock:        1,
				SupplierName: "OOO",
				SupplierInn:  "111111111111",
			},
			}}).Once()
		mpMock.On("GetMatchProducts").Return(service.MatchProducts{}).Once()
		dbMock.ExpectQuery("^select*").
			WillReturnRows(sqlResp.AddRow(
				"123",
				"addr",
				"",
				"",
				"",
				"man",
				"sup",
				"12222",
				"ПЕКТУСИН  ТАБ. Х10",
				0.0, 0.0, 0.0, 34080.0, "", "", ""))
		storeReciept[0].CreatedAt, storeReciept[0].UpdatedAt = time.Now().Format(time.RFC3339Nano), time.Now().Format(time.RFC3339Nano)
		mpMock.On("SendReceipt", mock2.AnythingOfType("[]store.Receipt")).Return()
		mpMock.On("SendOfdProducts", []service.Product{{Name: "2/5уп РУМАЛОН Р-Р В/М ВВЕД. АМП."}}, true, false).Return().Once()
		mpMock.On("SendOfdProducts", []service.Product(nil), false, true).Return().Once()
		unico.Parse()
		mpMock.AssertExpectations(t)
	})
}

func TestUnico_Manual(t *testing.T) {
	_ = godotenv.Overload("../../.env.test")
	config.Load()
	logger := windowsService.ConsoleLogger
	marketplace := &service.Marketplace{
		Revision: "3.0.1",
		Log:      logger,
		BaseURL:  config.Cfg.MarketplaceOptions.BaseUrl,
		Username: config.Cfg.MarketplaceOptions.Username,
		Password: config.Cfg.MarketplaceOptions.Password,
	}
	c := cache.New(5*time.Minute, 10*time.Minute)
	db, err := ConnectToErpDB(config.Cfg)
	assert.NoError(t, err)
	dt, err := time.Parse("02.01.2006", "02.02.2020")
	assert.NoError(t, err)
	unico := NewUnico(config.Cfg, marketplace, db, c, logger, dt)
	unico.Parse()
}
*/
