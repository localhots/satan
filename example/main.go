package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"

	"github.com/localhots/satan"
	"github.com/localhots/satan/example/daemons"
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "v", false, "Verbose mode")
	flag.Parse()
	if !debug {
		log.SetOutput(ioutil.Discard)
	}

	s := satan.Summon()
	s.AddDaemon(&daemons.NumberPrinter{})
	s.Start()
	defer s.Stop()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	for s := range sig {
		if s == os.Interrupt {
			return
		}
	}
}
