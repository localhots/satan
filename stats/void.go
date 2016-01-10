package stats

import (
	"time"
)

type Void struct{}

func (v *Void) Add(name string, dur time.Duration) {}

func (v *Void) Error(name string) {}
