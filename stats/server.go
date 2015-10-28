package stats

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	base

	history map[string][]*baseSnapshot
}

const (
	serverSnapshotIntervl = 5 * time.Second
	serverHistorySize     = 360 // 30 minutes of 5 second snapshots
)

func NewServer() *Server {
	s := &Server{}
	s.init()
	s.history = make(map[string][]*baseSnapshot)
	go s.takeSnapshots()

	return s
}

func (s *Server) History(rw http.ResponseWriter, _ *http.Request) {
	encoded, err := json.Marshal(s.history)
	if err != nil {
		http.Error(rw, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	rw.Header().Add("Access-Control-Allow-Origin", "*")
	rw.Write(encoded)
}

func (s *Server) takeSnapshots() {
	for range time.NewTicker(serverSnapshotIntervl).C {
		s.Lock()
		for name, stat := range s.stats {
			if len(s.history[name]) >= serverHistorySize {
				s.history[name] = s.history[name][1:]
			}
			s.history[name] = append(s.history[name], stat.snapshot())
		}
		s.Reset()
		s.Unlock()
	}
}
