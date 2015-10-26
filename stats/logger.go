package stats

import (
	"io"
	"os"
	"time"
)

type Logger struct {
	base

	out      io.Writer
	interval time.Duration
}

func NewLogger(out io.Writer, interval time.Duration) *Logger {
	l := &Logger{
		out:      out,
		interval: interval,
	}
	l.init()
	go l.printWithInterval()

	return l
}

func NewStdoutLogger(interval time.Duration) *Logger {
	return NewLogger(os.Stdout, interval)
}

func (l *Logger) Print() {
	for _, s := range l.stats {
		l.out.Write([]byte(s.String()))
		l.out.Write([]byte{'\n'})
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
