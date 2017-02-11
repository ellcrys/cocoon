package server

import (
	"net/http"

	"time"

	"github.com/gorilla/mux"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("http")

// Server defines a http server
// accepting commands from authorized admin
type Server struct {
}

// NewServer creates a new server
func NewServer() *Server {
	return new(Server)
}

// Start starts the http server
func (s *Server) Start(port string) {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", hello).Methods("GET")
	time.AfterFunc(time.Second*1, func() {
		log.Info("Started http server on port", port)
	})
	log.Fatal(http.ListenAndServe(":"+port, router))
}
