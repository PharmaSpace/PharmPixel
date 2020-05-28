package service

import (
	"Pixel/store"
	"fmt"
	"github.com/go-resty/resty/v2"
	serviceLib "github.com/kardianos/service"
	"github.com/thoas/go-funk"
	"log"
)

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

type Product struct {
	Name          string  `json:"name"`
	Manufacturer  string  `json:"manufacturer"`
	Export        bool    `json:"export"`
	PartNumber    string  `json:"partNumber"`
	Serial        string  `json:"serial"`
	Stock         float64 `json:"stock"`
	WarehouseName string  `json:"warehouseName"`
	SupplerName   string  `json:"supplerName"`
}

type MatchProducts struct {
	Data []struct {
		Export       bool   `json:"export"`
		ID           string `json:"id"`
		Manufacturer string `json:"manufacturer"`
		Name         string `json:"name"`
	} `json:"data"`
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
	resp, err := client.R().
		SetHeaders(m.getHeader()).
		SetResult(&matchProducts).
		Get(fmt.Sprintf("%s/api/v1/erp/products/match", m.BaseUrl))
	log.Printf("Статус получения мачинга %s", resp.Status())
	if resp.Status() != "200 OK" {
		log.Println(resp)
	}
	if err != nil {
		m.Log.Warningf("[WARNING] GetMatchProduct: %v", err)
	}
	return matchProducts
}
