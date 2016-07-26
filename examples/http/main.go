package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/localhots/shezmu"
	shezttp "github.com/localhots/shezmu/http"
)

func main() {
	sv := shezmu.Summon()
	server := shezttp.NewServer(sv, ":2255")
	server.Get("/", func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		w.Write([]byte("It works!"))
	})
	go server.Start()

	sv.StartDaemons()
	sv.HandleSignals()
}
