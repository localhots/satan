package daemons

import (
	"log"
	"math/rand"
	"time"

	"github.com/localhots/satan"
)

// NumberPrinter is a daemon that prints numbers once in a while.
type NumberPrinter struct {
	satan.BaseDaemon
}

// Startup sets up panic handler and starts enqueuing number printing jobs.
func (n *NumberPrinter) Startup() {
	n.HandlePanics(func() {
		log.Println("Oh, crap!")
	})

	n.SystemProcess(n.enqueue)
}

// Shutdown is empty due to the lack of cleanup.
func (n *NumberPrinter) Shutdown() {}

func (n *NumberPrinter) enqueue() {
	for n.Continue() {
		// Generate a random number between 1000 and 9999 and print it
		num := 1000 + rand.Intn(9000)
		n.Process(n.makeActor(num))

		// Sleep for a second or less
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	}
}

func (n *NumberPrinter) makeActor(num int) satan.Actor {
	return func() {
		// Making it crash sometimes
		if rand.Intn(20) == 0 {
			panic("Noooooooooo!")
		}

		log.Println("NumberPrinter says", num)
	}
}
