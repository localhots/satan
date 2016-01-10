package stats

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	base

	history map[string][]*serverStatsSnapshot
}

type serverStatsSnapshot struct {
	timestamp int64
	processed int64
	errors    int64
	min       float64
	p25       float64
	mean      float64
	median    float64
	p75       float64
	max       float64
}

const (
	// 60 of 10 second snapshots is 10 minutes worth of stats
	serverSnapshotIntervl = 10 * time.Second
	serverHistorySize     = 61 // +1 extra
)

func NewServer() *Server {
	s := &Server{}
	s.init()
	s.history = make(map[string][]*serverStatsSnapshot)
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
			s.history[name] = append(s.history[name], makeServerStatsSnapshot(stat))
		}
		s.Reset()
		s.Unlock()
	}
}

//
// Stats
//

func makeServerStatsSnapshot(s *baseStats) *serverStatsSnapshot {
	ps := s.time.Percentiles([]float64{0.25, 0.5, 0.75})

	return &serverStatsSnapshot{
		timestamp: time.Now().UTC().Unix(),
		processed: s.time.Count(),
		errors:    s.errors.Count(),
		min:       round(float64(s.time.Min())/1000000, 6),
		p25:       round(ps[0]/1000000, 6),
		mean:      round(s.time.Mean()/1000000, 6),
		median:    round(ps[1]/1000000, 6),
		p75:       round(ps[2]/1000000, 6),
		max:       round(float64(s.time.Max())/1000000, 6),
	}
}

// Implements json.Marshaler
func (s *serverStatsSnapshot) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("[%d,%d,%d,%.6f,%.6f,%.6f,%.6f,%.6f,%.6f]",
		s.timestamp, s.processed, s.errors, s.min, s.p25,
		s.mean, s.median, s.p75, s.max)), nil
}
