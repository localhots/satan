package server

import (
	"fmt"
	"net/http"

	"github.com/localhots/satan/stats"
)

type Server struct {
	port int
	ss   *stats.Server
	mux  *http.ServeMux
}

func New(port int, ss *stats.Server) *Server {
	return &Server{
		port: port,
		ss:   ss,
		mux:  http.NewServeMux(),
	}
}

func (s *Server) Start() {
	addr := fmt.Sprintf(":%d", s.port)
	s.mux.HandleFunc("/stats.json", s.ss.History)
	go http.ListenAndServe(addr, s.mux)
}
