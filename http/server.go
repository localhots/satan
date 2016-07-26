package http

import (
	"fmt"
	"net/http"

	hr "github.com/julienschmidt/httprouter"
	"github.com/localhots/shezmu"
)

// Server is an implementation of HTTP server that executes requests using
// Shezmu actors.
type Server struct {
	sv      *shezmu.Shezmu
	address string
	router  *hr.Router
}

type handler struct {
	shezmu.BaseDaemon

	name   string
	handle hr.Handle
}

// NewServer creates a new server.
func NewServer(sv *shezmu.Shezmu, address string) *Server {
	return &Server{
		sv:      sv,
		address: address,
		router:  hr.New(),
	}
}

// Start starts the server.
func (s *Server) Start() error {
	s.sv.Logger.Printf("Starting server at %s", s.address)
	return http.ListenAndServe(s.address, s.router)
}

// Get installs a handler for GET requests.
func (s *Server) Get(path string, handle hr.Handle) {
	s.router.GET(path, s.addHandlerDaemon(path, handle))
}

// Post installs a handler for POST requests.
func (s *Server) Post(path string, handle hr.Handle) {
	s.router.POST(path, s.addHandlerDaemon(path, handle))
}

// Put installs a handler for PUT requests.
func (s *Server) Put(path string, handle hr.Handle) {
	s.router.PUT(path, s.addHandlerDaemon(path, handle))
}

// Delete installs a handler for DELETE requests.
func (s *Server) Delete(path string, handle hr.Handle) {
	s.router.DELETE(path, s.addHandlerDaemon(path, handle))
}

func (s *Server) addHandlerDaemon(path string, handle hr.Handle) hr.Handle {
	h := &handler{
		name:   fmt.Sprintf("HTTP[%s]", path),
		handle: handle,
	}
	s.sv.AddDaemon(h)

	return h.process
}

func (h *handler) process(w http.ResponseWriter, r *http.Request, params hr.Params) {
	wait := make(chan struct{})
	h.Process(func() {
		defer close(wait)
		h.handle(w, r, params)
	})
	<-wait
}

func (h *handler) String() string {
	return h.name
}
