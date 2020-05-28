package model

import "encoding/json"

type KKT struct {
	Address  string `json:"address"`
	Kktregid string `json:"kktregid"`
}

type Documents struct {
	Count     int        `json:"count"`
	Documents []Document `json:"items"`
}

type Document struct {
	DateTime             int64  `json:"dateTime"`
	FiscalDocumentNumber int    `json:"fiscalDocumentNumber"`
	KktRegId             string `json:"kktRegId"`
	Nds20                int    `json:"nds20"`
	TotalSum             int    `json:"totalSum"`
	ProductName          string
	ProductQuantity      json.Number
	ProductPrice         int
	ProductTotalPrice    int
	Link                 string
	Ofd                  string
}
