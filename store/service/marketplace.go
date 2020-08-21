package service

import (
	"Pixel/store"
	"fmt"
	"github.com/go-resty/resty/v2"
	serviceLib "github.com/kardianos/service"
	"github.com/thoas/go-funk"
	"log"
)

type MarketPlaceInterface interface {
	SendProduct(products []Product)
	SendOfdProducts(products []Product, isOfd bool, isErp bool)
	SendReceipt(receipts []store.Receipt)
	SendReceiptN(receipts []store.ReceiptN)
	GetMatchProducts(filterDate string, isOfd bool, isErp bool) MatchProducts
}

type Marketpalce struct {
	Revision string
	Log      serviceLib.Logger
	BaseUrl  string
	Username string
	Password string
}

type DataProducts struct {
	Data []Product `json:"data"`
}

type DataReceipts struct {
	Data []store.Receipt `json:"data"`
}

type DataReceiptsN struct {
	Data []store.ReceiptN `json:"data"`
}

type OfdProductsRequest struct {
	Data  []Product `json:"data"`
	IsOfd bool      `json:"isOfd"`
	IsErp bool      `json:"isErp"`
}

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
	CreatedAt     string  `json:"createdAt"`
}

type MatchProductItem struct {
	Export       bool    `json:"export"`
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Stock        float64 `json:"stock"`
	SupplierName string  `json:"supplierName"`
	SupplierInn  string  `json:"supplierInn"`
}

type MatchProducts struct {
	Data []MatchProductItem `json:"data"`
}

type AuthSuccess struct {
	Data struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		UserID       string `json:"userId"`
	} `json:"data"`
}

var authSuccess AuthSuccess
var client *resty.Client

func init() {
	client = resty.New()
}

func (m *Marketpalce) getHeader() map[string]string {
	headers := map[string]string{}
	headers["authorization"] = authSuccess.Data.AccessToken
	headers["pixelVersion"] = m.Revision

	return headers
}

func (m *Marketpalce) auth() {
	headers := map[string]string{}
	headers["pixelVersion"] = m.Revision
	log.Println("auth")
	_, err := client.R().
		SetBody(fmt.Sprintf("{\"email\":\"%s\",\"password\":\"%s\"}", m.Username, m.Password)).
		SetHeaders(headers).
		SetResult(&authSuccess).
		Post(fmt.Sprintf("%s/api/v1/users/login", m.BaseUrl))
	if err != nil {
		m.Log.Errorf("[ERROR] auth: %v", err)
	}
}

func (m *Marketpalce) SendProduct(products []Product) {
	m.auth()
	log.Printf("SendProduct count: %d", len(products))
	prChunk := funk.Chunk(products, 1000)
	if val, ok := prChunk.([][]Product); ok {
		for _, prs := range val {
			dataProduct := DataProducts{Data: prs}
			resp, err := client.R().
				SetHeaders(m.getHeader()).
				SetBody(dataProduct).
				Post(fmt.Sprintf("%s/api/v1/erp/products", m.BaseUrl))
			log.Printf("Статус отправки продуктов %s", resp.Status())
			if resp.Status() != "200 OK" {
				log.Println(resp)
			}
			if err != nil {
				m.Log.Warningf("[WARNING] SendProduct: %v", err)
			}
		}
	}
}

func (m *Marketpalce) SendOfdProducts(products []Product, isOfd bool, isErp bool) {
	m.auth()
	if isErp {
		log.Printf("SendErpProduct count: %d", len(products))
	} else if isOfd {
		log.Printf("SendOfdProduct count: %d", len(products))
	}

	prChunk := funk.Chunk(products, 1000)
	if val, ok := prChunk.([][]Product); ok {
		for _, prs := range val {
			dataProduct := OfdProductsRequest{Data: prs, IsOfd: isOfd, IsErp: isErp}
			resp, err := client.R().
				SetHeaders(m.getHeader()).
				SetBody(dataProduct).
				Post(fmt.Sprintf("%s/api/v1/erp/products/ofd", m.BaseUrl))
			log.Printf("Статус отправки ofd продуктов %s", resp.Status())
			if resp.Status() != "200 OK" {
				log.Println(resp)
			}
			if err != nil {
				m.Log.Warningf("[WARNING] SendProduct: %v", err)
			}
		}
	}
}

func (m *Marketpalce) SendReceipt(receipts []store.Receipt) {
	m.auth()
	err := m.Log.Infof("SendReceipts Valid count: %d", len(receipts))
	dataReceipts := DataReceipts{Data: receipts}
	resp, err := client.R().
		SetHeaders(m.getHeader()).
		SetBody(dataReceipts).
		Post(fmt.Sprintf("%s/api/v1/erp/sales/valid", m.BaseUrl))
	log.Printf("Статус отправки заказов %s", resp.Status())
	if resp.Status() != "200 OK" {
		log.Println(resp)
	}
	if err != nil {
		m.Log.Warningf("[WARNING] SendReceipts: %v", err)
	}
}

func (m *Marketpalce) SendReceiptN(receipts []store.ReceiptN) {
	m.auth()
	err := m.Log.Infof("SendReceipts count: %d", len(receipts))
	dataReceipts := DataReceiptsN{Data: receipts}
	resp, err := client.R().
		SetHeaders(m.getHeader()).
		SetBody(dataReceipts).
		Post(fmt.Sprintf("%s/api/v1/erp/sales", m.BaseUrl))
	log.Printf("Статус отправки заказов %s", resp.Status())
	if resp.Status() != "200 OK" {
		log.Println(resp)
	}
	if err != nil {
		m.Log.Warningf("[WARNING] SendReceiptsN: %v", err)
	}
}

func (m *Marketpalce) GetMatchProduct() MatchProducts {
	m.auth()
	matchProducts := MatchProducts{}
	url := fmt.Sprintf("%s/api/v1/erp/products/match", m.BaseUrl)
	resp, err := client.R().
		SetHeaders(m.getHeader()).
		SetResult(&matchProducts).
		Get(url)
	log.Printf("Статус получения мачинга %s", resp.Status())
	if resp.Status() != "200 OK" {
		log.Println(resp)
	}
	if err != nil {
		m.Log.Warningf("[WARNING] GetMatchProduct: %v", err)
	}
	return matchProducts
}

func (m *Marketpalce) GetMatchProducts(filterDate string, isOfd bool, isErp bool) MatchProducts {
	m.auth()
	matchProducts := MatchProducts{}
	url := fmt.Sprintf("%s/api/v1/erp/products/ofd?filterDate=%s&isOfd=%t&isErp=%t", m.BaseUrl, filterDate, isOfd, isErp)
	resp, err := client.R().
		SetHeaders(m.getHeader()).
		SetResult(&matchProducts).
		Get(url)
	log.Printf("Статус получения мачинга %s", resp.Status())
	if resp.Status() != "200 OK" {
		log.Println(resp)
	}
	if err != nil {
		m.Log.Warningf("[WARNING] GetMatchProduct: %v", err)
	}
	return matchProducts
}
