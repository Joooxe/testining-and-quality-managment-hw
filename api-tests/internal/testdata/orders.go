package testdata

import (
	"crypto/rand"
	"fmt"

	"github.com/NZRV-88/TestProject/api-tests/internal/orders"
)

const (
	ExpectedTotal    = 32
	CancellationNote = "Ошибочный заказ"
)

func NewOrder() orders.NewOrder {
	suffix := rand.Text()
	apartment := "12"
	return orders.NewOrder{
		Address: orders.Address{
			City:      "Amsterdam",
			Street:    "Damrak " + suffix,
			House:     "1",
			Apartment: &apartment,
			Postcode:  "1012",
		},
		Items: []orders.Item{
			{Name: "Margherita " + suffix, Size: "MEDIUM", Qty: 2},
			{Name: "Pepperoni " + suffix, Size: "LARGE", Qty: 1},
		},
	}
}

func NewUUID() string {
	var id [16]byte
	rand.Read(id[:])
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", id[:4], id[4:6], id[6:8], id[8:10], id[10:])
}
