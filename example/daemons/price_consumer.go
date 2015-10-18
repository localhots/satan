package daemons

import (
	"log"

	"github.com/localhots/satan"
)

// PriceConsumer consumes price update messages and prints them to the console.
type PriceConsumer struct {
	satan.BaseConsumer
}

// PriceUpdate describes a price update message.
type PriceUpdate struct {
	Product string  `json:"product"`
	Amount  float64 `json:"amount"`
}

// Startup creates a new subscription for ProductPriceUpdates topic.
func (p *PriceConsumer) Startup() {
	b.Subscribe("ProductPriceUpdates", func(u PriceUpdate) {
		log.Printf("Price for %q is now $%.2f", u.Product, u.Amount)
	})
	p.LimitRate(5, 1*time.Second)
}

// Shutdown is empty because PriceConsumer requires no cleanup upon exiting.
func (p *PriceConsumer) Shutdown() {}
