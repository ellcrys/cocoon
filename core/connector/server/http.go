package server

import (
	"io"
	"net/http"
	"time"

	"strings"

	logging "github.com/op/go-logging"
)

var httpLog = logging.MustGetLogger("connector.http")

// HTTPServer defines an http server to
// handle http related functionalities
type HTTPServer struct {
}

// NewHTTPServer creates a new instance of HTTPServer
func NewHTTPServer() *HTTPServer {
	return &HTTPServer{}
}

// Start starts the server
func (server *HTTPServer) Start(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", server.HealthCheck)
	time.AfterFunc(2*time.Second, func() {
		httpLog.Infof("Start HTTP server on port %s", strings.Split(addr, ":")[1])
	})
	http.ListenAndServe(addr, mux)
}

// HealthCheck handles heath checks by external services
func (server *HTTPServer) HealthCheck(w http.ResponseWriter, r *http.Request) {
	httpLog.Info("Received health check query")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "alive")
}
