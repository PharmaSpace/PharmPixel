package model

type Receipt struct {
	PharmacyID      string
	PharmacyAddress string
	Date            string
	KKM             string
	InvoiceNumber   string
	Manufacturer    string
	Supplier        string
	SupplierINN     string
	Name            string
	PriceWoVat      string
	PriceWVat       string
	Vat             string
	TotalPrice      string
	TotalNumber     string
	ShipmentNumber  string
	Series          string
}

