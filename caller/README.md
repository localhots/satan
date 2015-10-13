# Caller

Package caller is used to dynamicly call functions with data unmarshalled
into the functions' first argument. It's main purpose is to hide common
unmarshalling code from each function's implementation thus reducing
boilerplate and making the code sexier.

[Documentation](https://godoc.org/github.com/localhots/uberdaemon/caller)

```go
package main

import (
    "github.com/localhots/uberdaemon/caller"
)

type message struct {
    Title string `json:"title"`
    Body  string `json:"body"`
}

func processMessage(m message) {
    fmt.Printf("Title: %s\nBody: %s\n", m.Title, m.Body)
}

func main() {
    c, _ := caller.New(processMessage)
    c.Call(`{"title": "Hello", "body": "World"}`)
}
```
