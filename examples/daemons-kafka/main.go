package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/localhots/shezmu"
	"github.com/localhots/shezmu/examples/daemons-kafka/daemons"
	"github.com/localhots/shezmu/examples/daemons-kafka/kafka"
	"github.com/localhots/shezmu/server"
	"github.com/localhots/shezmu/stats"
)

func main() {
	var brokers string

	flag.StringVar(&brokers, "brokers", "127.0.0.1:9092", "Kafka broker addresses separated by space")
	flag.Parse()

	kafka.Initialize(strings.Split(brokers, " "))
	defer kafka.Shutdown()

	statsLogger := stats.NewStdoutLogger(0)
	defer statsLogger.Print()

	statsServer := stats.NewServer()
	server := server.New(6464, statsServer)
	server.Start()

	s := shezmu.Summon()
	s.DaemonStats = stats.NewGroup(statsLogger, statsServer)

	s.AddDaemon(&daemons.NumberPrinter{})
	s.AddDaemon(&daemons.PriceConsumer{})

	s.StartDaemons()
	defer s.StopDaemons()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGHUP)
	switch <-sig {
	case syscall.SIGHUP:
		s.StopDaemons()
		s.StartDaemons()
	case syscall.SIGINT:
		return
	}
}
