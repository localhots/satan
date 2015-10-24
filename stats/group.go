package stats

import (
	"time"

	"github.com/localhots/satan"
)

type Group struct {
	backends []satan.StatsPublisher
}

func NewGroup(backends ...satan.StatsPublisher) *Group {
	return &Group{
		backends: backends,
	}
}

func (g *Group) Add(name string, dur time.Duration) {
	for _, b := range g.backends {
		b.Add(name, dur)
	}
}

func (g *Group) Error(name string) {
	for _, b := range g.backends {
		b.Error(name)
	}
}
