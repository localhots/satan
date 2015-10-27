package satan

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/localhots/satan/stats"
)

// Satan is the master daemon.
type Satan struct {
	SubscribeFunc SubscribeFunc
	Publisher     Publisher
	DaemonStats   stats.Publisher
	Logger        *log.Logger
	NumWorkers    int

	daemons      []Daemon
	queue        chan *task
	runtimeStats stats.Manager

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
	DefaultNumWorkers = 100
)

// Summon creates a new instance of Satan.
func Summon() *Satan {
	return &Satan{
		Logger:          log.New(os.Stdout, "[daemons] ", log.LstdFlags),
		NumWorkers:      DefaultNumWorkers,
		queue:           make(chan *task),
		runtimeStats:    stats.NewBasicStats(),
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
	base.logger = s.Logger
	base.shutdown = s.shutdownSystem

	go d.Startup()
	s.daemons = append(s.daemons, d)
}

// StartDaemons starts all registered daemons.
func (s *Satan) StartDaemons() {
	s.Logger.Printf("Starting %d workers", s.NumWorkers)
	for i := 0; i < s.NumWorkers; i++ {
		go s.runWorker()
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

	fmt.Println(s.runtimeStats.Fetch(stats.Latency))
}

func (s *Satan) runWorker() {
	s.wgWorkers.Add(1)
	defer s.wgWorkers.Done()
	defer func() {
		if err := recover(); err != nil {
			s.Logger.Printf("Worker crashed. Error: %v\n", err)
			debug.PrintStack()
			go s.runWorker() // Restarting worker
		}
	}()

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
	s.runtimeStats.Add(stats.Latency, time.Duration(dur))

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
			s.Logger.Printf("System task %s recovered from a panic\nError: %v\n", t, err)
			debug.PrintStack()

			t.createdAt = time.Now()
			s.queue <- t // Restarting task
		} else {
			s.Logger.Printf("System task %s has stopped\n", t)
		}
	}()

	s.Logger.Printf("Starting system task %s\n", t)
	t.actor() // <--- ACTION STARTS HERE
}

func (s *Satan) processGeneralTask(t *task) {
	defer func() {
		if err := recover(); err != nil {
			if s.DaemonStats != nil {
				s.DaemonStats.Error(t.daemon.base().String())
			}
			t.daemon.base().handlePanic(err)
			s.Logger.Printf("Daemon %s recovered from a panic\nError: %v\n", t.daemon.base(), err)
			debug.PrintStack()
		}
	}()
	if s.DaemonStats != nil {
		defer func(start time.Time) {
			dur := time.Now().UnixNano() - start.UnixNano()
			s.DaemonStats.Add(t.daemon.base().String(), time.Duration(dur))
		}(time.Now())
	}

	t.actor() // <--- ACTION STARTS HERE
}

func (t *task) String() string {
	if t.name == "" {
		return fmt.Sprintf("[unnamed %s process]", t.daemon)
	}

	return fmt.Sprintf("%s[%s]", t.daemon, t.name)
}
