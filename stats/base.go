package stats

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
)

type Manager interface {
	Publisher
	Fetch(name string) Stats
	Reset()
}

type Publisher interface {
	Add(name string, dur time.Duration)
	Error(name string)
}

type Stats interface {
	Processed() int64
	Errors() int64
	Min() int64
	Mean() float64
	P95() float64
	Max() int64
	StdDev() float64
}

const (
	DefaultSampleSize = 1000
	Latency           = "Latency"
	TaskWait          = "TaskWait"
)

//
// base
//

type base struct {
	sync.Mutex
	stats      map[string]*baseStats
	sampleSize int
}

func (b *base) Add(name string, dur time.Duration) {
	b.metrics(name).time.Update(int64(dur))
}

func (b *base) Error(name string) {
	b.metrics(name).errors.Inc(1)
}

func (b *base) Fetch(name string) Stats {
	return b.metrics(name)
}

func (b *base) Reset() {
	for _, s := range b.stats {
		s.time.Clear()
		s.errors.Clear()
	}
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

		if b.sampleSize == 0 {
			b.sampleSize = DefaultSampleSize
		}
		b.stats[name] = &baseStats{
			name:   name,
			time:   metrics.NewHistogram(metrics.NewUniformSample(b.sampleSize)),
			errors: metrics.NewCounter(),
		}
	}

	return b.stats[name]
}

//
// baseStats
//

type baseStats struct {
	name   string
	time   metrics.Histogram
	errors metrics.Counter
}

func (s *baseStats) Processed() int64 {
	return s.time.Count()
}

func (s *baseStats) Errors() int64 {
	return s.errors.Count()
}

func (s *baseStats) Min() int64 {
	return s.time.Min()
}

func (s *baseStats) Mean() float64 {
	return s.time.Mean()
}

func (s *baseStats) P95() float64 {
	return s.time.Percentile(0.95)
}

func (s *baseStats) Max() int64 {
	return s.time.Max()
}

func (s *baseStats) StdDev() float64 {
	return s.time.StdDev()
}

func (s *baseStats) String() string {
	return fmt.Sprintf("%s statistics:\n"+
		"Processed: %10d\n"+
		"Errors:    %10d\n"+
		"Min:       %10s\n"+
		"Mean:      %10s\n"+
		"95%%:       %10s\n"+
		"Max:       %10s\n"+
		"StdDev:    %10s",
		s.name,
		s.time.Count(),
		s.errors.Count(),
		formatDuration(float64(s.time.Min())),
		formatDuration(s.time.Mean()),
		formatDuration(s.time.Percentile(0.95)),
		formatDuration(float64(s.time.Max())),
		formatDuration(s.time.StdDev()),
	)
}

//
// Helpers
//

func formatDuration(dur float64) string {
	switch {
	case dur < 1000:
		return fmt.Sprintf("%10.0fns", dur)
	case dur < 1000000:
		return fmt.Sprintf("%10.3fμs", dur/1000)
	case dur < 1000000000:
		return fmt.Sprintf("%10.3fms", dur/1000000)
	default:
		return fmt.Sprintf("%10.3fs", dur/1000000000)
	}
}

func round(num float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return float64(int(num*pow+0.5)) / pow
}
