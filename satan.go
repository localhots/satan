package satan

import (
	"log"
	"sync"
	"time"
)

// Satan is the master daemon.
type Satan struct {
	SubscribeFunc SubscribeFunc
	Publisher     Publisher

	daemons  []Daemon
	queue    chan *task
	shutdown chan struct{}
	wg       sync.WaitGroup
	latency  *statistics
}

// Actor is a function that could be executed by daemon workers.
type Actor func()

// SubscribeFunc is a function that is used by daemons to subscribe to messages.
type SubscribeFunc func(consumer, topic string) Streamer

// Streamer is the interface that wraps message consumers. Error handling
// should be provided by the implementation. Feel free to panic.
type Streamer interface {
	Messages() <-chan []byte
	Close()
}

// Publisher is the interface that wraps message publishers. Error handling
// should be provided by the implementation. Feel free to panic.
type Publisher interface {
	Publish(msg []byte)
	Close()
}

const (
	defaultNumWorkers = 10
)

// Summon creates a new instance of Satan.
func Summon() *Satan {
	return &Satan{
		queue:    make(chan *task),
		latency:  newStatistics(),
		shutdown: make(chan struct{}),
	}
}

// AddDaemon adds a new daemon.
func (s *Satan) AddDaemon(d Daemon) {
	base := d.base()
	base.self = d
	base.subscribeFunc = s.SubscribeFunc
	base.publisher = s.Publisher
	base.queue = s.queue
	base.shutdown = make(chan struct{})
	base.stats = newStatistics()

	go d.Startup()
	s.daemons = append(s.daemons, d)
}

// StartDaemons starts all registered daemons.
func (s *Satan) StartDaemons() {
	s.wg.Add(defaultNumWorkers)
	for i := 0; i < defaultNumWorkers; i++ {
		go func(i int) {
			s.runWorker(i)
			s.wg.Done()
		}(i)
	}
}

// StopDaemons stops all running daemons.
func (s *Satan) StopDaemons() {
	for _, d := range s.daemons {
		close(d.base().shutdown)
		d.Shutdown()

		stats := d.base().stats.snapshot()
		log.Printf("%s daemon performace statistics:\n%s\n", d.base(), stats)
	}

	close(s.shutdown)
	s.wg.Wait()
	close(s.queue)

	log.Printf("Task processing latency statistics:\n%s\n", s.latency.snapshot())
}

func (s *Satan) runWorker(i int) {
	log.Printf("Starting worker #%d", i+1)
	defer log.Printf("Worker #%d has stopped", i+1)

	for {
		select {
		case t := <-s.queue:
			dur := time.Now().UnixNano() - t.createdAt.UnixNano()
			s.latency.add(time.Duration(dur))
			if restart := t.process(); restart {
				s.queue <- t
			}
		default:
			select {
			case <-s.shutdown:
				return
			default:
			}
		}
	}
}
