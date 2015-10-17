package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/localhots/satan"
	"github.com/localhots/satan/example/daemons"
	"github.com/localhots/satan/example/kafka"
)

func main() {
	var debug bool
	var brokers string

	flag.BoolVar(&debug, "v", false, "Verbose mode")
	flag.StringVar(&brokers, "brokers", "127.0.0.1:9092", "Kafka broker addresses separated by space")
	flag.Parse()

	log.SetOutput(ioutil.Discard)
	if debug {
		log.SetOutput(os.Stderr)
	}

	kafka.Initialize(strings.Split(brokers, " "))
	defer kafka.Shutdown()

	s := satan.Summon()
	s.SubscribeFunc = kafka.Subscribe
	s.AddDaemon(&daemons.NumberPrinter{})
	s.AddDaemon(&daemons.PriceConsumer{})

	s.StartDaemons()
	defer s.StopDaemons()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	<-sig
}
