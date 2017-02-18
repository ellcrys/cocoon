package server

import (
	"fmt"
	"net/http"

	"time"

	"github.com/gorilla/mux"
	logging "github.com/op/go-logging"
)

var httpLog = logging.MustGetLogger("http")

// HTTPServer defines a http server
// accepting commands from authorized admin
type HTTPServer struct {
}

// NewHTTPServer creates a new server
func NewHTTPServer() *HTTPServer {
	return new(HTTPServer)
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Hello")
}

// Start starts the http server
func (s *HTTPServer) Start(port string) {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", hello).Methods("GET")
	time.AfterFunc(time.Second*1, func() {
		httpLog.Info("Started http server on port %s", port)
	})
	httpLog.Fatal(http.ListenAndServe(":"+port, router))
}
