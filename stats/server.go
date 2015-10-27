package stats

import (
	"encoding/json"
	"net/http"
)

type Server struct {
	base
}

func NewServer() *Server {
	s := &Server{}
	s.init()

	return s
}

func (s *Server) ServeHTTP(rw http.ResponseWriter, _ *http.Request) {
	stats := make(map[string]map[string]interface{})
	for name, stat := range s.stats {
		stats[name] = map[string]interface{}{
			"processed": stat.time.Count(),
			"errors":    stat.errors.Count(),
			"min":       float64(stat.time.Min()) / 1000000,
			"mean":      stat.time.Mean() / 1000000,
			"95%":       stat.time.Percentile(0.95) / 1000000,
			"max":       float64(stat.time.Max()) / 1000000,
			"stddev":    stat.time.StdDev() / 1000000,
		}
	}

	encoded, err := json.MarshalIndent(stats, "", "    ")
	if err != nil {
		panic(err)
	}

	rw.Write(encoded)
}
