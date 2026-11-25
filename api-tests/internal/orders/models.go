package orders

import "encoding/json"

type Address struct {
	City      string  `json:"city"`
	Street    string  `json:"street"`
	House     string  `json:"house"`
	Apartment *string `json:"apartment,omitempty"`
	Postcode  string  `json:"postcode"`
}

type Item struct {
	Name string `json:"name"`
	Size string `json:"size"`
	Qty  int    `json:"qty"`
}

type NewOrder struct {
	Address Address `json:"address"`
	Items   []Item  `json:"items"`
}

type Summary struct {
	ID           string      `json:"id"`
	Status       string      `json:"status"`
	TotalAmount  json.Number `json:"totalAmount"`
	Currency     string      `json:"currency"`
	DeliveryZone string      `json:"deliveryZone"`
}

type Order struct {
	Summary
	Address            Address `json:"address"`
	Items              []Item  `json:"items"`
	CancellationReason *string `json:"cancellationReason"`
}

type APIError struct {
	Status int    `json:"status"`
	Path   string `json:"path"`
}
