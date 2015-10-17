package satan

import (
	"fmt"
	"log"
	"runtime/debug"
	"time"
)

type task struct {
	daemon    Daemon
	actor     Actor
	createdAt time.Time
	system    bool
	name      string
}

func (t *task) process() (restart bool) {
	defer func(start time.Time) {
		dur := time.Now().UnixNano() - start.UnixNano()
		t.daemon.base().stats.add(time.Duration(dur))

		if err := recover(); err != nil {
			if t.system {
				log.Printf("System process %s recovered from a panic\nError: %v\n", t, err)
				debug.PrintStack()
				restart = true
			} else {
				t.daemon.base().handlePanic(err)
			}
		}
	}(time.Now())

	t.actor() // <--- THE ACTION HAPPENS HERE
	return
}

func (t *task) String() string {
	if t.name == "" {
		return fmt.Sprintf("[unnamed %s process]", t.daemon.base())
	}

	return fmt.Sprintf("%s[%s]", t.daemon.base(), t.name)
}
