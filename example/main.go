package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"

	"github.com/localhots/uberdaemon"
	"github.com/localhots/uberdaemon/example/daemons"
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "v", false, "Verbose mode")
	flag.Parse()
	if !debug {
		log.SetOutput(ioutil.Discard)
	}

	uberd := uberdaemon.New()
	uberd.AddDaemon(&daemons.NumberPrinter{})
	uberd.Start()
	defer uberd.Stop()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	for s := range sig {
		if s == os.Interrupt {
			return
		}
	}
}
