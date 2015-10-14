package satan

import (
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"time"
)

// Daemon is the interface that contains a set of methods required to be
// implemented in order to be treated as a daemon.
type Daemon interface {
	// Startup implementation should:
	//
	// func (d *DaemonName) Startup() {
	//     // 1. Set up a panic handler
	//     b.HandlePanics(func() {
	//         log.Error("Oh, crap!")
	//     })
	//
	//     // 2. If the daemon is also a consumer we need to subscribe for
	//     // topics that would be consumed by the daemon
	//     b.Subscribe("ProductPriceUpdates", func(p PriceUpdate) {
	// 	       log.Printf("Price for %q is now $%.2f", p.Product, p.Amount)
	//     })
	//
	//     // 3. If the daemon is doing some IO it is a good idea to limit the
	//     // rate of its execution
	//     b.LimitRate(10, 1 * time.Second)
	// }
	Startup()

	// Shutdown implementation should clean up all daemon related stuff:
	// close channels, process the last batch of items, etc.
	Shutdown()

	// base is a (hack) function that allows the Daemon interface to reference
	// underlying BaseDaemon structure.
	base() *BaseDaemon
}

// Actor is a function that could be executed by daemon workers.
type Actor func()

// BaseDaemon is the parent structure for all daemons.
type BaseDaemon struct {
	self         Daemon
	name         string
	stats        *statistics
	enqueue      func(*task)
	panicHandler func()
	shutdown     chan struct{}
}

// Process creates a task and then adds it to processing queue.
func (b *BaseDaemon) Process(a Actor) {
	b.enqueue(&task{
		daemon:    b.self,
		actor:     a,
		createdAt: time.Now(),
	})
}

// HandlePanics sets up a panic handler function for the daemon.
func (b *BaseDaemon) HandlePanics(f func()) {
	b.panicHandler = f
}

// ShutdownRequested returns a channel that is closed the moment daemon shutdown
// is requested.
func (b *BaseDaemon) ShutdownRequested() <-chan struct{} {
	return b.shutdown
}

// ShouldShutdown returns true if daemon should shutdown and false otherwise.
func (b *BaseDaemon) ShouldShutdown() bool {
	return b.shutdown == nil
}

// String returns the name of the Deamon unerlying struct.
func (b *BaseDaemon) String() string {
	if b.name == "" {
		b.name = strings.Split(fmt.Sprintf("%T", b.self), ".")[1]
	}

	return b.name
}

// initialize saves a reference to the child daemon which is then used to print
// the daemons' name. It also initializes other struct fields.
func (b *BaseDaemon) initialize(self Daemon, enqueue func(*task)) {
	b.self = self
	b.stats = newStatistics()
	b.enqueue = enqueue
	b.shutdown = make(chan struct{})
}

// base is a (hack) function that allows the Daemon interface to reference
// underlying BaseDaemon structure.
func (b *BaseDaemon) base() *BaseDaemon {
	return b
}

func (b *BaseDaemon) handlePanic() {
	if err := recover(); err != nil {
		b.stats.registerError()
		if b.panicHandler != nil {
			b.panicHandler()
		}
		log.Printf("Daemon %s recovered from panic. Error: %v\n", b, err)
		debug.PrintStack()
	}
}
