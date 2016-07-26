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
		Logger:          log.New(os.Stdout, "", log.LstdFlags),
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
	base.queue = s.queue
	base.logger = s.Logger
	base.shutdown = s.shutdownSystem

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

	s.Logger.Println("Setting up daemons")
	for _, d := range s.daemons {
		s.setupDaemon(d)
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

	// Re-open closed channels to allow starting new deamons afterwards
	s.shutdownSystem = make(chan struct{})
	s.shutdownWorkers = make(chan struct{})
	s.queue = make(chan *task)

	fmt.Println(s.runtimeStats.Fetch(stats.Latency))
}

func (s *Shezmu) setupDaemon(d Daemon) {
	defer func() {
		if err := recover(); err != nil {
			s.Logger.Printf("Failed to setup daemon %s due to process termination", d)
		}
	}()

	s.queue <- &task{
		daemon:    d,
		actor:     d.Startup,
		createdAt: time.Now(),
		system:    true,
		name:      "startup",
	}
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
			s.Logger.Printf("System task %s finished\n", t)
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
	return fmt.Sprintf("%s[%s]", t.daemon, t.name)
}
