# Caller

Package caller is used to dynamically call functions with data unmarshalled
into the functions' first argument. Its main purpose is to hide common
unmarshalling code from each function implementation thus reducing
boilerplate and making package interaction code sexier.

[Documentation](https://godoc.org/github.com/localhots/satan/caller)

### Example

```go
package main

import (
    "log"
    "github.com/localhots/satan/caller"
    "github.com/path/to/package/messenger"
)

type PriceUpdate struct {
    Product string  `json:"product"`
    Amount  float32 `json:"amount"`
}

func main() {
    messenger.Subscribe("ProductPriceUpdates", func(p PriceUpdate) {
        log.Printf("Price for %q is now $%.2f", p.Product, p.Amount)
    })
    messenger.Deliver()
}
```

Support code:

```go
package messenger

import (
    "github.com/localhots/satan/caller"
)

type item struct {
    topic   string
    payload []byte
}

var queue <-chan item
var subscriptions = make(map[string][]*caller.Caller)

func Subscribe(topic string, callback interface{}) {
    c, err := caller.New(processMessage)
    if err != nil {
        panic(err)
    }
    subcriptions[topic] = append(subcriptions[topic], c)
}

func Deliver() {
    for itm := range queue {
        for _, c := range subscriptions[itm.topic] {
            // Payload example:
            // {"product": "Paperclip", "amount": 0.01}
            c.Call(itm.payload)
        }
    }
}
```
