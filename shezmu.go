package shezmu

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/localhots/shezmu/stats"
)

// Shezmu is the master daemon.
type Shezmu struct {
	Subscriber  Subscriber
	Publisher   Publisher
	DaemonStats stats.Publisher
	Logger      Logger
	NumWorkers  int

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

// Subscriber is the interface that is used by daemons to subscribe to messages.
type Subscriber interface {
	Subscribe(consumerName, topic string) Streamer
}

// Streamer is the interface that wraps message consumers. Error handling
// should be provided by the implementation. Feel free to panic.
type Streamer interface {
	Messages() <-chan []byte
	Close()
}

// Publisher is the interface that wraps message publishers. Error handling
// should be provided by the implementation. Feel free to panic.
type Publisher interface {
	Publish(topic string, msg []byte, meta interface{})
	Close()
}

// Logger is the interface that implements minimal logging functions.
type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type task struct {
	daemon    Daemon
	actor     Actor
	createdAt time.Time
	system    bool
	name      string
}

const (
	// DefaultNumWorkers is the default number of workers that would process
	// tasks.
	DefaultNumWorkers = 100
)

// Summon creates a new instance of Shezmu.
func Summon() *Shezmu {
	return &Shezmu{
		DaemonStats:     &stats.Void{},
		Logger:          log.New(os.Stdout, "[daemons] ", log.LstdFlags),
		NumWorkers:      DefaultNumWorkers,
		queue:           make(chan *task),
		runtimeStats:    stats.NewBasicStats(),
		shutdownWorkers: make(chan struct{}),
		shutdownSystem:  make(chan struct{}),
	}
}

// AddDaemon adds a new daemon.
func (s *Shezmu) AddDaemon(d Daemon) {
	base := d.base()
	base.self = d
	base.subscriber = s.Subscriber
	base.publisher = s.Publisher
	base.queue = s.queue
	base.logger = s.Logger
	base.shutdown = s.shutdownSystem

	go d.Startup()
	s.daemons = append(s.daemons, d)
}

// ClearDaemons clears the list of added daemons. StopDaemons() function MUST be
// called before calling ClearDaemons().
func (s *Shezmu) ClearDaemons() {
	s.daemons = []Daemon{}
}

// StartDaemons starts all registered daemons.
func (s *Shezmu) StartDaemons() {
	s.Logger.Printf("Starting %d workers", s.NumWorkers)
	for i := 0; i < s.NumWorkers; i++ {
		go s.runWorker()
	}
}

// StopDaemons stops all running daemons.
func (s *Shezmu) StopDaemons() {
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

func (s *Shezmu) runWorker() {
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

func (s *Shezmu) processTask(t *task) {
	dur := time.Now().Sub(t.createdAt)
	s.runtimeStats.Add(stats.Latency, dur)

	if t.system {
		s.processSystemTask(t)
	} else {
		s.processGeneralTask(t)
	}
}

func (s *Shezmu) processSystemTask(t *task) {
	// Abort starting a system task if shutdown was already called. Otherwise
	// incrementing a wait group counter will cause a panic. This should be an
	// extremely rare scenario when a system task crashes and tries to restart
	// after a shutdown call.
	select {
	case <-s.shutdownSystem:
		return
	default:
	}

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

func (s *Shezmu) processGeneralTask(t *task) {
	defer func() {
		if err := recover(); err != nil {
			s.DaemonStats.Error(t.daemon.String())
			t.daemon.base().handlePanic(err)
			s.Logger.Printf("Daemon %s recovered from a panic\nError: %v\n", t.daemon, err)
			debug.PrintStack()
		}
	}()
	defer func(start time.Time) {
		dur := time.Now().Sub(start)
		s.DaemonStats.Add(t.daemon.String(), dur)
	}(time.Now())

	t.actor() // <--- ACTION STARTS HERE
}

func (t *task) String() string {
	if t.name == "" {
		return fmt.Sprintf("[unnamed %s process]", t.daemon)
	}

	return fmt.Sprintf("%s[%s]", t.daemon, t.name)
}
