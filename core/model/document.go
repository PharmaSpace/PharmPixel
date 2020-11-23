package model

// KKT структура кассы
type KKT struct {
	Address  string `json:"address"`
	Kktregid string `json:"kktregid"`
}

// Documents структура документов
type Documents struct {
	Count     int        `json:"count"`
	Documents []Document `json:"items"`
}

// Document структура документа
type Document struct {
	DateTime              int64  `json:"dateTime"`
	FiscalDocumentNumber  int    `json:"fiscalDocumentNumber"`
	KktRegID              string `json:"kktRegId"`
	Nds20                 int    `json:"nds20"`
	TotalSum              int    `json:"totalSum"`
	ProductName           string
	ProductQuantity       int
	ProductPrice          int
	ProductTotalPrice     int
	Link                  string
	Ofd                   string
	FiscalDocumentNumber2 string // FP, DocNumber и т.п - доп. значения по к-ым можно найти совпадение в чеках
	FiscalDocumentNumber3 string // FP, DocNumber и т.п - доп. значения по к-ым можно найти совпадение в чеках
}
