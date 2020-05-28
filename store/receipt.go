package store

type Receipt struct {
	DateTime             string `json:"dateTime" db:"dateTime"`
	FiscalDocumentNumber int    `json:"fiscalDocumentNumber" db:"fiscalDocumentNumber"`
	Inn                  string `json:"inn" db:"inn"`
	KktRegId             string `json:"kktRegId" db:"kktRegId"`
	Link                 string `json:"link" db:"link"`
	Name                 string `json:"name" db:"name"`
	Ofd                  string `json:"ofd" db:"ofd"`
	Price                int    `json:"price" db:"price"`
	PriceSellIn          int    `json:"priceSellIn" db:"priceSellIn"`
	ProductId            string `json:"productId" db:"productId"`
	Quantity             string `json:"quantity" db:"quantity"`
	Total                int    `json:"total" db:"total"`
	TotalSum             int    `json:"totalSum" db:"totalSum"`
	SupplerName          string `json:"supplerName"`
	PointName            string `json:"pointName"`
	Series               string `json:"series"`
	CreatedAt            string `json:"createdAt" db:"createdAt" `
	UpdatedAt            string `json:"updatedAt" db:"updatedAt"`
}

type ReceiptN struct {
	ID           string `json:"id"`
	DatePay      string `json:"datePay"`
	Manufacture  string `json:"manufacture"`
	Name         string `json:"name"`
	Number       string `json:"number"`
	PointName    string `json:"pointName"`
	PriceSellIn  int    `json:"priceSellIn"`
	PriceSellOut int    `json:"priceSellOut"`
	ProductId    string `json:"productId"`
	Provider     string `json:"provider"`
	Quantity     string `json:"quantity"`
	SupplerName  string `json:"supplerName"`
	Series       string `json:"series"`
}

func (r *Receipt) PrepareUntrusted() {
	r.ProductId = ""
}
