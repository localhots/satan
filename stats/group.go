package stats

import (
	"time"
)

type Group struct {
	backends []Publisher
}

func NewGroup(backends ...Publisher) *Group {
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
