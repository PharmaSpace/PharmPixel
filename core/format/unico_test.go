package format

import (
	"Pixel/config"
	"Pixel/core/format/mock"
	"Pixel/store/service"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/joho/godotenv"
	windowsService "github.com/kardianos/service"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
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
		cache := cache.New(5*time.Minute, 10*time.Minute)
		setCache(cache, ofdItems)
		mpMock := &mock.MarketPlaceMock{}
		dbMock := &mock.DBMock{}
		date := time.Now()

		unico := Unico(config.Cfg, mpMock, dbMock, cache, logger, date)
		unico.Config.UnicoOptions.SqlDriver = "test"
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
		cache := cache.New(5*time.Minute, 10*time.Minute)
		setCache(cache, ofdItems)
		mpMock := &mock.MarketPlaceMock{}
		//dbMock := &mock.DBMock{}
		db, dbMock, _ := sqlmock.New()
		date := time.Now()
		unico := Unico(config.Cfg, mpMock, db, cache, logger, date)
		unico.Config.UnicoOptions.SqlDriver = "test"
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
		cache := cache.New(5*time.Minute, 10*time.Minute)
		setCache(cache, ofdItems)
		mpMock := &mock.MarketPlaceMock{}
		//dbMock := &mock.DBMock{}
		db, dbMock, _ := sqlmock.New()
		date := time.Now()
		unico := Unico(config.Cfg, mpMock, db, cache, logger, date)
		unico.Config.UnicoOptions.SqlDriver = "test"
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

func TestUnico_Manual (t *testing.T) {
	_ = godotenv.Overload("../../.env.test")
	config.Load()
	logger := windowsService.ConsoleLogger
	marketplace := &service.Marketpalce{
		Revision: "3.0.1",
		Log:      logger,
		BaseUrl:  config.Cfg.MarketplaceOptions.BaseUrl,
		Username: config.Cfg.MarketplaceOptions.Username,
		Password: config.Cfg.MarketplaceOptions.Password,
	}
	cache := cache.New(5*time.Minute, 10*time.Minute)
	db, err := ConnectToErpDB(config.Cfg)
	assert.NoError(t, err)
	dt, err := time.Parse("02.01.2006", "02.02.2020")
	assert.NoError(t, err)
	unico := Unico(config.Cfg, marketplace, db, cache, logger, dt)
	unico.Parse()
}


