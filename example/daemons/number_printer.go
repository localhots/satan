package daemons

import (
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
	n.HandlePanics(func(err interface{}) {
		n.Logf("Oh, crap! There was a panic, take a look: %v", err)
	})

	n.LimitRate(3, time.Second)
	n.SystemProcess("Random Number Generator", n.generateNumbers)
}

// Shutdown is empty due to the lack of cleanup.
func (n *NumberPrinter) Shutdown() {}

func (n *NumberPrinter) generateNumbers() {
	for n.Continue() {
		if rand.Intn(7) == 0 {
			panic("Number generator don't work on Sundays!")
		}
		// Generate a random number between 1000 and 9999 and print it
		num := 1000 + rand.Intn(9000)
		n.Process(n.makeActor(num))
	}
}

func (n *NumberPrinter) makeActor(num int) satan.Actor {
	return func() {
		// Making it crash sometimes
		if rand.Intn(10) == 0 {
			panic("Nooooo! Random number generator returned a zero!")
		}

		n.Log("Number printer says:", num)
	}
}
