package model

import "time"

// Product структура продкута
type Product struct {
	PharmacyID      string
	PharmacyAddress string
	Name            string
	Supplier        string
	SupplierINN     string
	Manufacturer    string
	CountryOfOrigin string
	ShipmentNumber  string
	Series          string
	Inventory       float64
	EAN             string
	CreatedAt       time.Time
}
