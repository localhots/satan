package satan

import (
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"time"

	"github.com/localhots/satan/caller"
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

// BaseDaemon is the parent structure for all daemons.
type BaseDaemon struct {
	subscribeFunc SubscribeFunc
	publisher     Publisher
	self          Daemon
	name          string
	stats         *statistics
	queue         chan<- *task
	panicHandler  func()
	shutdown      chan struct{}
}

var (
	errMissingSubscriptionFun = errors.New("subscription function is not set up")
	errMissingPublisher       = errors.New("publisher is not set up")
)

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

// Subscribe subscriasdsdfsdgdfgdfsg sdgsdfg sdfgs dfgdfgdfg.
func (d *BaseDaemon) Subscribe(topic string, fun interface{}) {
	d.SystemProcess(func() {
		if d.subscribeFunc == nil {
			panic(errMissingSubscriptionFun)
		}

		stream := d.subscribeFunc(d.String(), topic)
		defer stream.Close()

		cf, err := caller.New(fun)
		if err != nil {
			panic(err)
		}

		for {
			select {
			case msg := <-stream.Messages():
				d.Process(func() { cf.Call(msg) })
			case <-d.shutdown:
				return
			}
		}
	})
}

// Publish sends a message to the publisher.
func (d *BaseDaemon) Publish(msg []byte) {
	if d.publisher == nil {
		panic(errMissingPublisher)
	}

	d.publisher.Publish(msg)
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

// Continue returns true if daemon should proceed and false if it should stop.
func (d *BaseDaemon) Continue() bool {
	select {
	case <-d.shutdown:
		return false
	default:
		return true
	}
}

// String returns the name of the Deamon unerlying struct.
func (d *BaseDaemon) String() string {
	if d.name == "" {
		d.name = strings.Split(fmt.Sprintf("%T", d.self), ".")[1]
	}

	return d.name
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
