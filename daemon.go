package shezmu

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/juju/ratelimit"
	"github.com/localhots/shezmu/caller"
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
	//     // 2. If the daemon is doing some IO it is a good idea to limit the
	//     // rate of its execution
	//     b.LimitRate(10, 1 * time.Second)
	//
	//     // 3. If the daemon is also a consumer we need to subscribe for
	//     // topics that would be consumed by the daemon
	//     b.Subscribe("ProductPriceUpdates", func(p PriceUpdate) {
	// 	       log.Printf("Price for %q is now $%.2f", p.Product, p.Amount)
	//     })
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
	self         Daemon
	name         string
	queue        chan<- *task
	logger       Logger
	panicHandler PanicHandler
	subscriber   Subscriber
	publisher    Publisher
	shutdown     chan struct{}
	limit        *ratelimit.Bucket
}

// PanicHandler is a function that handles panics. Duh!
type PanicHandler func(interface{})

var (
	errMissingSubscriber = errors.New("subscriber is not set up")
	errMissingPublisher  = errors.New("publisher is not set up")
)

// Process creates a task and then adds it to processing queue.
func (d *BaseDaemon) Process(a Actor) {
	if d.limit != nil {
		d.limit.Wait(1)
	}
	d.queue <- &task{
		daemon:    d.self,
		actor:     a,
		createdAt: time.Now(),
	}
}

// SystemProcess creates a system task that is restarted in case of failure
// and then adds it to processing queue.
func (d *BaseDaemon) SystemProcess(name string, a Actor) {
	d.queue <- &task{
		daemon:    d.self,
		actor:     a,
		createdAt: time.Now(),
		system:    true,
		name:      name,
	}
}

// Subscribe subscriasdsdfsdgdfgdfsg sdgsdfg sdfgs dfgdfgdfg.
func (d *BaseDaemon) Subscribe(topic string, fun interface{}) {
	name := fmt.Sprintf("subscription for topic %q", topic)
	d.SystemProcess(name, func() {
		if d.subscriber == nil {
			panic(errMissingSubscriber)
		}

		stream := d.subscriber.Subscribe(d.String(), topic)
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

// LimitRate limits the daemons' processing rate.
func (d *BaseDaemon) LimitRate(times int, per time.Duration) {
	rate := float64(time.Second) / float64(per) * float64(times)
	if rate <= 0 {
		d.Logf("Daemon %s processing rate was limited to %d. Using 1 instead", d.base(), rate)
		rate = 1.0
	}
	d.Logf("Daemon %s processing rate is limited to %.2f ops/s", d.base(), rate)
	d.limit = ratelimit.NewBucketWithRate(rate, 1)
}

// HandlePanics sets up a panic handler function for the daemon.
func (d *BaseDaemon) HandlePanics(f PanicHandler) {
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

// Log logs values using shezmu.Logger.Println function.
func (d *BaseDaemon) Log(v ...interface{}) {
	if d.logger != nil {
		d.logger.Println(v...)
	}
}

// Logf logs values using shezmu.Logger.Printf function.
func (d *BaseDaemon) Logf(format string, v ...interface{}) {
	if d.logger != nil {
		d.logger.Printf(format, v...)
	}
}

// Shutdown is the empty implementation of the daemons' Shutdown function that
// is inherited and used by default.
func (d *BaseDaemon) Shutdown() {}

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

func (d *BaseDaemon) handlePanic(err interface{}) {
	if d.panicHandler != nil {
		d.panicHandler(err)
	}
}
