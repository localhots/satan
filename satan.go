package satan

import (
	"log"
	"sync"
	"time"

	"github.com/localhots/satan/backend"
	"github.com/localhots/satan/backend/kafka"
)

// Satan is the master daemon.
type Satan struct {
	daemons  []Daemon
	backend  backend.Backend
	queue    chan *task
	wg       sync.WaitGroup
	latency  *statistics
	shutdown chan struct{}
}

// Actor is a function that could be executed by daemon workers.
type Actor func()

type task struct {
	daemon    Daemon
	actor     Actor
	createdAt time.Time
	system    bool
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
	d.initialize(d, s.queue)
	if c, ok := d.(Consumer); ok {
		c.setBackend(s.backend)
	}
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
	close(s.shutdown)
	s.wg.Wait()
	close(s.queue)
	s.backend.Close()

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

			// log.Printf("Daemon #%d got some job to do!", i+1)
			s.processTask(t)
		default:
			select {
			case <-s.shutdown:
				return
			default:
			}
		}
	}
}

func (s *Satan) processTask(t *task) {
	defer t.daemon.base().handlePanic()
	start := time.Now()

	t.actor()

	dur := time.Now().UnixNano() - start.UnixNano()
	t.daemon.base().stats.add(time.Duration(dur))
}

// InitializeKafka initializes Kafka backend.
func (s *Satan) InitializeKafka(id string, brokers []string) error {
	k, err := kafka.New(id, brokers)
	if err != nil {
		return err
	}
	if err = k.InitializeProducer(); err != nil {
		return err
	}
	if err = k.InitializeConsumer(); err != nil {
		return err
	}

	s.backend = k
	return nil
}
