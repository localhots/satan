package satan

import (
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"time"
)

// Satan is the master daemon.
type Satan struct {
	SubscribeFunc SubscribeFunc
	Publisher     Publisher

	daemons []Daemon
	queue   chan *task
	latency *statistics

	wgWorkers       sync.WaitGroup
	wgSystem        sync.WaitGroup
	shutdownWorkers chan struct{}
	shutdownSystem  chan struct{}
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

type task struct {
	daemon    Daemon
	actor     Actor
	createdAt time.Time
	system    bool
	name      string
}

const (
	defaultNumWorkers = 10
)

// Summon creates a new instance of Satan.
func Summon() *Satan {
	return &Satan{
		queue:           make(chan *task),
		latency:         newStatistics(),
		shutdownWorkers: make(chan struct{}),
		shutdownSystem:  make(chan struct{}),
	}
}

// AddDaemon adds a new daemon.
func (s *Satan) AddDaemon(d Daemon) {
	base := d.base()
	base.self = d
	base.subscribeFunc = s.SubscribeFunc
	base.publisher = s.Publisher
	base.queue = s.queue
	base.shutdown = s.shutdownSystem
	base.stats = newStatistics()

	go d.Startup()
	s.daemons = append(s.daemons, d)
}

// StartDaemons starts all registered daemons.
func (s *Satan) StartDaemons() {
	for i := 0; i < defaultNumWorkers; i++ {
		go func(i int) {
			s.runWorker(i)
		}(i)
	}
}

// StopDaemons stops all running daemons.
func (s *Satan) StopDaemons() {
	close(s.shutdownSystem)
	for _, d := range s.daemons {
		d.Shutdown()
	}

	s.wgSystem.Wait()
	close(s.shutdownWorkers)
	s.wgWorkers.Wait()
	close(s.queue)

	for _, d := range s.daemons {
		stats := d.base().stats.snapshot()
		log.Printf("%s daemon performace statistics:\n%s\n", d.base(), stats)
	}
	log.Printf("Task processing latency statistics:\n%s\n", s.latency.snapshot())
}

func (s *Satan) runWorker(i int) {
	s.wgWorkers.Add(1)
	defer s.wgWorkers.Done()
	log.Printf("Starting worker #%d", i+1)
	defer log.Printf("Worker #%d has stopped", i+1)

	for {
		select {
		case t := <-s.queue:
			s.processTask(t)
		case <-s.shutdownWorkers:
			return
		}
	}
}

func (s *Satan) processTask(t *task) {
	dur := time.Now().UnixNano() - t.createdAt.UnixNano()
	s.latency.add(time.Duration(dur))

	if t.system {
		s.processSystemTask(t)
	} else {
		s.processGeneralTask(t)
	}
}

func (s *Satan) processSystemTask(t *task) {
	s.wgSystem.Add(1)
	defer s.wgSystem.Done()
	defer func() {
		if err := recover(); err != nil {
			log.Printf("System task %s recovered from a panic\nError: %v\n", t, err)
			debug.PrintStack()

			t.createdAt = time.Now()
			s.queue <- t // Restarting task
		} else {
			log.Printf("System task %s has stopped\n", t)
		}
	}()

	log.Printf("Starting system task %s\n", t)
	t.actor() // <--- THE ACTION HAPPENS HERE
}

func (s *Satan) processGeneralTask(t *task) {
	defer t.daemon.base().handlePanic()
	defer func(start time.Time) {
		dur := time.Now().UnixNano() - start.UnixNano()
		t.daemon.base().stats.add(time.Duration(dur))
	}(time.Now())

	t.actor() // <--- THE ACTION HAPPENS HERE
}

func (t *task) String() string {
	if t.name == "" {
		return fmt.Sprintf("[unnamed %s process]", t.daemon.base())
	}

	return fmt.Sprintf("%s[%s]", t.daemon.base(), t.name)
}
