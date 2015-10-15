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

	// initialize is also a hack that is used by the Satan to initialize
	// base daemon fields.
	initialize(self Daemon, queue chan<- *task)
}

// BaseDaemon is the parent structure for all daemons.
type BaseDaemon struct {
	self         Daemon
	name         string
	stats        *statistics
	queue        chan<- *task
	panicHandler func()
	shutdown     chan struct{}
}

// Process creates a task and then adds it to processing queue.
func (d *BaseDaemon) Process(a Actor) {
	d.enqueue(a, false)
}

// SystemProcess creates a system task that is restarted in case of failure
// and then adds it to processing queue.
func (d *BaseDaemon) SystemProcess(a Actor) {
	d.enqueue(a, true)
}

func (d *BaseDaemon) enqueue(a Actor, system bool) {
	d.queue <- &task{
		daemon:    d.self,
		actor:     a,
		createdAt: time.Now(),
		system:    system,
	}
}

// HandlePanics sets up a panic handler function for the daemon.
func (d *BaseDaemon) HandlePanics(f func()) {
	d.panicHandler = f
}

// ShutdownRequested returns a channel that is closed the moment daemon shutdown
// is requested.
func (d *BaseDaemon) ShutdownRequested() <-chan struct{} {
	return d.shutdown
}

// ShouldShutdown returns true if daemon should shutdown and false otherwise.
func (d *BaseDaemon) ShouldShutdown() bool {
	select {
	case <-d.shutdown:
		return true
	default:
		return false
	}
}

// String returns the name of the Deamon unerlying struct.
func (d *BaseDaemon) String() string {
	if d.name == "" {
		d.name = strings.Split(fmt.Sprintf("%T", d.self), ".")[1]
	}

	return d.name
}

// initialize saves a reference to the child daemon which is then used to print
// the daemons' name. It also initializes other struct fields.
func (d *BaseDaemon) initialize(self Daemon, queue chan<- *task) {
	d.self = self
	d.stats = newStatistics()
	d.queue = queue
	d.shutdown = make(chan struct{})
}

// base is a (hack) function that allows the Daemon interface to reference
// underlying BaseDaemon structure.
func (d *BaseDaemon) base() *BaseDaemon {
	return d
}

func (d *BaseDaemon) handlePanic() {
	if err := recover(); err != nil {
		d.stats.registerError()
		if d.panicHandler != nil {
			d.panicHandler()
		}
		log.Printf("Daemon %s recovered from panic. Error: %v\n", d, err)
		debug.PrintStack()
	}
}
