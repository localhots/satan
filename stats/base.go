package stats

import (
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
)

type base struct {
	sync.Mutex
	stats map[string]*baseStats
}

type baseStats struct {
	time   metrics.Histogram
	errors metrics.Counter
}

func (b *base) Add(name string, dur time.Duration) {
	b.metrics(name).time.Update(int64(dur))
}

func (b *base) Error(name string) {
	b.metrics(name).errors.Inc(1)
}

func (b *base) init() {
	b.stats = make(map[string]*baseStats)
}

func (b *base) metrics(name string) *baseStats {
	if _, ok := b.stats[name]; !ok {
		b.Lock()
		defer b.Unlock()

		// Double checking being protected by mutex
		if s, ok := b.stats[name]; ok {
			return s
		}

		b.stats[name] = &baseStats{
			time:   metrics.NewHistogram(metrics.NewUniformSample(1000)),
			errors: metrics.NewCounter(),
		}
	}

	return b.stats[name]
}
