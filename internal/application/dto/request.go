package dto

import (
	"github.com/victus-est-deus/shipment/internal/domain/valueobject"
)

type CreateShipmentRequest struct {
	ReferenceNumber        string
	Origin                 string
	Destination            string
	DriverName             string
	DriverPhone            string
	UnitNumber             string
	ShipmentAmount         float64
	ShipmentCurrency       string
	DriverRevenue          float64
	DriverRevenueCurrency  string
}

type UpdateStatusRequest struct {
	ShipmentID string
	Status     string
	Location   string
	Notes      string
}

func (r CreateShipmentRequest) ToShipmentAmount() (valueobject.Money, error) {
	currency := r.ShipmentCurrency
	if currency == "" {
		currency = "USD"
	}
	return valueobject.NewMoney(r.ShipmentAmount, currency)
}

func (r CreateShipmentRequest) ToDriverRevenue() (valueobject.Money, error) {
	currency := r.DriverRevenueCurrency
	if currency == "" {
		currency = "USD"
	}
	return valueobject.NewMoney(r.DriverRevenue, currency)
}
