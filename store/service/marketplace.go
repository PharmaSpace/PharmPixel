package service

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	serviceLib "github.com/kardianos/service"
	"github.com/thoas/go-funk"
	"log"
	"os"
	"pixel/sentry"
	"pixel/store"
)

// MarketPlaceInterface структура
type MarketPlaceInterface interface {
	SendOfdProducts(products []Product, isOfd, isErp bool) error
	SendReceipt(receipts []store.Receipt) error
	GetMatchProducts(filterDate string, isOfd, isErp bool) (MatchProducts, error)
}

// Marketplace структура для подключение к марктеплейсу
type Marketplace struct {
	Revision string
	Log      serviceLib.Logger
	BaseURL  string
	Username string
	Password string
}

// DataProducts структура для продуктов
type DataProducts struct {
	Data []Product `json:"data"`
}

// DataReceipts структура для чеков
type DataReceipts struct {
	Data []store.Receipt `json:"data"`
}

// OfdProductsRequest создание прокудтов
type OfdProductsRequest struct {
	Data  []Product `json:"data"`
	IsOfd bool      `json:"isOfd"`
	IsErp bool      `json:"isErp"`
}

// Product структура продукта
type Product struct {
	Name          string  `json:"name"`
	Manufacturer  string  `json:"manufacturer"`
	Export        bool    `json:"export"`
	PartNumber    string  `json:"partNumber"`
	Serial        string  `json:"serial"`
	Stock         float64 `json:"stock"`
	WarehouseName string  `json:"warehouseName"`
	SupplerName   string  `json:"supplerName"`
	SupplerInn    string  `json:"supplerInn"`
	CreatedAt     string  `json:"date"`
}

// MatchProductItem получение смаченных продуктов
type MatchProductItem struct {
	Export       bool    `json:"export"`
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Stock        float64 `json:"stock"`
	SupplierName string  `json:"supplierName"`
	SupplierInn  string  `json:"supplierInn"`
}

// MatchProducts список запросов
type MatchProducts struct {
	Data       []MatchProductItem `json:"data"`
	Pagination Pagination         `json:"pagination"`
}

type Pagination struct {
	TotalCount  int32 `json:"totalCount"`
	TotalPages  int32 `json:"totalPages"`
	CurrentPage int32 `json:"currentPage"`
	Size        int32 `json:"size"`
}

// Auth автризация
type Auth struct {
	Data struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		UserID       string `json:"userId"`
	} `json:"data"`
	Code    string
	Message string
	Status  int
}

type Merchant struct {
	MerchantID string `json:"merchantId"`
	PointID    string `json:"pointId"`
}

var authCredentials Auth
var client *resty.Client

//nolint:gochecknoinits // can't avoid it in this place
func init() {
	client = resty.New()
}

func (m *Marketplace) getHeader() map[string]string {
	headers := map[string]string{}
	headers["authorization"] = authCredentials.Data.AccessToken
	headers["pixelVersion"] = m.Revision

	return headers
}

func (m *Marketplace) auth() error {
	headers := map[string]string{}
	headers["pixelVersion"] = m.Revision
	_, err := client.SetRetryWaitTime(10000).R().
		SetBody(fmt.Sprintf("{\"email\":\"%s\",\"password\":\"%s\"}", m.Username, m.Password)).
		SetHeaders(headers).
		SetResult(&authCredentials).
		Post(fmt.Sprintf("%s/api/v1/users/login", m.BaseURL))
	if err != nil || authCredentials.Status != 200 {
		sentry.NewSentry(m.Username, "", m.Revision).Error(err)
		err = m.Log.Errorf("[ERROR] auth: %v, code:%s, message:%s", err, authCredentials.Code, authCredentials.Message)
		if err != nil {
			log.Printf("[ERROR] auth: %v, code:%s, message:%s", err, authCredentials.Code, authCredentials.Message)
			return err
		}
		os.Exit(1)
	}
	return nil
}

// WhoAmi получение информации по пользователю
func (m *Marketplace) WhoAmi() (*Merchant, error) {
	err := m.auth()
	if err != nil {
		return nil, err
	}
	var merchant Merchant
	resp, _ := client.R().SetHeaders(m.getHeader()).SetResult(&merchant).Get(fmt.Sprintf("%s/api/v1/erp/whoami", m.BaseURL))
	if resp != nil {
		fmt.Println(resp)
	}
	return &merchant, nil
}

// SendOfdProducts отправка продуктов
func (m *Marketplace) SendOfdProducts(products []Product, isOfd, isErp bool) error {
	err := m.auth()
	if err != nil {
		return err
	}
	prChunk := funk.Chunk(products, 1000)
	if val, ok := prChunk.([][]Product); ok {
		for _, prs := range val {
			dataProduct := OfdProductsRequest{Data: prs, IsOfd: isOfd, IsErp: isErp}
			resp, err := client.SetRetryWaitTime(10000).R().
				SetHeaders(m.getHeader()).
				SetBody(dataProduct).
				Post(fmt.Sprintf("%s/api/v1/erp/products/ofd", m.BaseURL))
			if err != nil {
				err = m.Log.Warningf("[WARNING] SendProduct: %v", err)
				if err != nil {
					log.Printf("[WARNING] SendProduct: %v", err)
				}
				return err
			}
			log.Printf("Статус отправки %d продуктов %s", len(prs), resp.Status())
			if resp.Status() != "200 OK" {
				log.Println(resp)
				return fmt.Errorf(resp.String())
			}

		}
	}
	return nil
}

// SendReceipt отправка чеков
func (m *Marketplace) SendReceipt(receipts []store.Receipt) error {
	var err error
	err = m.auth()
	if err != nil {
		return err
	}
	err = m.Log.Infof("SendReceipts Valid count: %d", len(receipts))
	if err != nil {
		return err
	}
	dataReceipts := DataReceipts{Data: receipts}
	resp, err := client.SetRetryWaitTime(10000).R().
		SetHeaders(m.getHeader()).
		SetBody(dataReceipts).
		Post(fmt.Sprintf("%s/api/v1/erp/sales/valid", m.BaseURL))
	if err != nil {
		err = m.Log.Warningf("[WARNING] SendReceipts: %v", err)
		return err
	}
	log.Printf("Статус отправки заказов %s", resp.Status())
	if resp.Status() != "200 OK" {
		log.Println(resp)
		return fmt.Errorf(resp.String())
	}
	return nil
}

// GetMatchProducts получение продуктов с фильтром по дате
func (m *Marketplace) GetMatchProducts(filterDate string, isOfd, isErp bool) (matchProducts MatchProducts, err error) {
	err = m.auth()
	if err != nil {
		return matchProducts, err
	}
	log.Printf("Старт получения мачинга")
	totalPage := int32(1)
	for page := int32(1); page <= totalPage; page++ {
		var matchProduct MatchProducts
		url := fmt.Sprintf("%s/api/v1/erp/products/ofd?filterDate=%s&isOfd=%t&isErp=%t&page=%d&size=%d", m.BaseURL, filterDate, isOfd, isErp, page, 100)
		var resp *resty.Response
		resp, err = client.SetRetryWaitTime(10000).R().
			SetHeaders(m.getHeader()).
			SetResult(&matchProduct).
			Get(url)
		if err != nil {
			err = m.Log.Warningf("[WARNING] GetMatchProduct: %v", err)
			return matchProducts, err
		}

		if resp.Status() != "200 OK" {
			log.Println(resp)
			return matchProducts, fmt.Errorf(resp.String())
		}
		totalPage = matchProduct.Pagination.TotalPages
		matchProducts.Data = append(matchProducts.Data, matchProduct.Data...)
	}
	log.Printf("Получили товаров мачинга: %d", len(matchProducts.Data))
	return matchProducts, err
}
