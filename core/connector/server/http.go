package server

import (
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/ncodes/cocoon/core/config"
	logging "github.com/op/go-logging"
)

var httpLog *logging.Logger

// HTTP defines a structure for an HTTP server
// that provides REST API services.
type HTTP struct {
	rpc *RPC
}

// NewHTTP creates an new http server instance
func NewHTTP(rpc *RPC) *HTTP {
	httpLog = config.MakeLogger("connector.http", "")
	return &HTTP{rpc}
}

// getRouter returns the router
func (s *HTTP) getRouter() *mux.Router {
	r := mux.NewRouter()
	g := r.PathPrefix("/v1").Subrouter()
	g.HandleFunc("/invoke", s.invokeCocoonCode)
	return r
}

// Start starts the http server. Passes true to the startedCh channel
// when started
func (s *HTTP) Start(addr string, startedCh chan bool) error {
	time.AfterFunc(2*time.Second, func() {
		httpLog.Infof("Started HTTP API server @ %s", addr)
		startedCh <- true
	})
	err := http.ListenAndServe(addr, s.getRouter())
	return err
}

func (s *HTTP) invokeCocoonCode(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello world!")
}
