package satan

import (
	"log"
	"sync"
	"time"
)

// Satan is the master daemon.
type Satan struct {
	daemons  []Daemon
	queue    chan *task
	wg       sync.WaitGroup
	latency  *statistics
	shutdown chan struct{}
}

type task struct {
	daemon    Daemon
	actor     Actor
	createdAt time.Time
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
	d.base().initialize(d, s.queue)
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

		log.Printf("%s daemon performace statistics:\n%s\n",
			d.base(), d.base().stats.snapshot())
	}
	close(s.queue)
	s.wg.Wait()

	log.Printf("Task processing latency statistics:\n%s\n", s.latency.snapshot())
}

func (s *Satan) runWorker(i int) {
	log.Printf("Starting worker #%d", i+1)
	defer log.Printf("Worker #%d has stopped", i+1)

	for t := range s.queue {
		dur := time.Now().UnixNano() - t.createdAt.UnixNano()
		s.latency.add(time.Duration(dur))

		log.Printf("Daemon #%d got some job to do!", i+1)
		s.processTask(t)
	}
}

func (s *Satan) processTask(t *task) {
	defer t.daemon.base().handlePanic()
	start := time.Now()

	t.actor()

	dur := time.Now().UnixNano() - start.UnixNano()
	t.daemon.base().stats.add(time.Duration(dur))
}
