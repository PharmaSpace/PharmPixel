package store

// Receipt структура чека
type Receipt struct {
	DateTime             string `json:"dateTime" db:"dateTime"`
	FiscalDocumentNumber int    `json:"fiscalDocumentNumber" db:"fiscalDocumentNumber"`
	Inn                  string `json:"inn" db:"inn"`
	KktRegID             string `json:"kktRegId" db:"kktRegId"`
	Link                 string `json:"link" db:"link"`
	Name                 string `json:"name" db:"name"`
	Ofd                  string `json:"ofd" db:"ofd"`
	Price                int    `json:"price" db:"price"`
	PriceSellIn          int    `json:"priceSellIn" db:"priceSellIn"`
	ProductID            string `json:"productId" db:"productId"`
	Quantity             string `json:"quantity" db:"quantity"`
	Total                int    `json:"total" db:"total"`
	TotalSum             int    `json:"totalSum" db:"totalSum"`
	SupplerName          string `json:"supplerName"`
	SupplerInn           string `json:"supplerInn"`
	PointName            string `json:"pointName"`
	Series               string `json:"series"`
	IsValidated          bool   `json:"isValidated"`
	CreatedAt            string `json:"createdAt" db:"createdAt" `
	UpdatedAt            string `json:"updatedAt" db:"updatedAt"`
}

// PrepareUntrusted подготовка чека
func (r *Receipt) PrepareUntrusted() {
	r.ProductID = ""
}
