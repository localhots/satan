package server

import (
	"fmt"
	"net/http"
)

type Server struct {
	port int
	mux  *http.ServeMux
}

func NewServer(port int) *Server {
	return &Server{
		port: port,
		mux:  http.NewServeMux(),
	}
}

func (s *Server) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

func (s *Server) Start() {
	addr := fmt.Sprintf(":%d", port)
	go http.ListenAndServe(addr, s.mux)
}
