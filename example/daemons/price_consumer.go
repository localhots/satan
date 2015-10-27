package daemons

import (
	"time"

	"github.com/localhots/satan"
)

// PriceConsumer consumes price update messages and prints them to the console.
type PriceConsumer struct {
	satan.BaseDaemon
}

// PriceUpdate describes a price update message.
type PriceUpdate struct {
	Product string  `json:"product"`
	Amount  float64 `json:"amount"`
}

// Startup creates a new subscription for ProductPriceUpdates topic.
func (p *PriceConsumer) Startup() {
	p.Subscribe("ProductPriceUpdates", func(u PriceUpdate) {
		p.Logf("Price for %q is now $%.2f", u.Product, u.Amount)
	})
	p.LimitRate(5, time.Second)
}
