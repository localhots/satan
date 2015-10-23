package stats

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
)

type Logger struct {
	sync.Mutex
	out      io.Writer
	interval time.Duration
	stats    map[string]*loggerStats
}

type loggerStats struct {
	time   metrics.Histogram
	errors metrics.Counter
}

func NewLogger(out io.Writer, interval time.Duration) *Logger {
	l := &Logger{
		out:      out,
		interval: interval,
		stats:    make(map[string]*loggerStats),
	}
	go l.printWithInterval()

	return l
}

func NewStdoutLogger(interval time.Duration) *Logger {
	l := &Logger{
		out:      os.Stdout,
		interval: interval,
		stats:    make(map[string]*loggerStats),
	}
	go l.printWithInterval()

	return l
}

func (l *Logger) Add(name string, dur time.Duration) {
	l.metrics(name).time.Update(int64(dur))
}

func (l *Logger) Error(name string) {
	l.metrics(name).errors.Inc(1)
}

func (l *Logger) Print() {
	for name, s := range l.stats {
		fmt.Fprintf(l.out, "%s statistics:\n"+
			"Processed: %d\n"+
			"Errors:    %d\n"+
			"Min:       %.8fms\n"+
			"Max:       %.8fms\n"+
			"95%%:       %.8fms\n"+
			"Mean:      %.8fms\n"+
			"StdDev:    %.8fms\n",
			name,
			s.time.Count(),
			s.errors.Count(),
			float64(s.time.Min())/1000000,
			float64(s.time.Max())/1000000,
			s.time.Percentile(0.95)/1000000,
			s.time.Mean()/1000000,
			s.time.StdDev()/1000000,
		)
		s.time.Clear()
		s.errors.Clear()
	}
}

func (l *Logger) printWithInterval() {
	if l.interval == 0 {
		return
	}

	for range time.NewTicker(l.interval).C {
		l.Print()
	}
}

func (l *Logger) metrics(name string) *loggerStats {
	if _, ok := l.stats[name]; !ok {
		l.Lock()
		defer l.Unlock()

		// Double checking being protected by mutex
		if s, ok := l.stats[name]; ok {
			return s
		}

		l.stats[name] = &loggerStats{
			time:   metrics.NewHistogram(metrics.NewUniformSample(1000)),
			errors: metrics.NewCounter(),
		}
	}

	return l.stats[name]
}
