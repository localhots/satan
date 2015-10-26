package stats

import (
	"testing"
	"time"
)

func TestNewGroup(t *testing.T) {
	g := NewGroup(newGroupItemMock(), newGroupItemMock())

	if bc := len(g.backends); bc != 2 {
		t.Fatalf("Expected backends to contain 2 items, got %d", bc)
	}
}

func TestGroupAdd(t *testing.T) {
	g := NewGroup(newGroupItemMock(), newGroupItemMock())

	dur := 5 * time.Second
	g.Add("", dur)

	for _, m := range g.backends {
		select {
		case dur1 := <-m.(*groupItemMock).addCalls:
			if dur1 != dur {
				t.Errorf("Expected a 5s duration, got %d", dur1)
			}
		default:
			t.Error("Mock item didn't receive an Add call")
		}
	}
}

func TestGroupError(t *testing.T) {
	g := NewGroup(newGroupItemMock(), newGroupItemMock())

	g.Error("")

	for _, m := range g.backends {
		select {
		case <-m.(*groupItemMock).errorCalls:
		default:
			t.Error("Mock item didn't receive an Error call")
		}
	}
}

//
// Mock
//

type groupItemMock struct {
	addCalls   chan time.Duration
	errorCalls chan struct{}
}

func (g *groupItemMock) Add(_ string, dur time.Duration) {
	g.addCalls <- dur
}

func (g *groupItemMock) Error(_ string) {
	g.errorCalls <- struct{}{}
}

func newGroupItemMock() *groupItemMock {
	return &groupItemMock{
		addCalls:   make(chan time.Duration, 1),
		errorCalls: make(chan struct{}, 1),
	}
}
