package server

import (
	"encoding/json"
	"net/http"
	"time"

	"strings"

	"net/url"

	"github.com/asaskevich/govalidator"
	"github.com/ellcrys/util"
	"github.com/gorilla/mux"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/connector/server/proto_connector"
	logging "github.com/op/go-logging"
	context "golang.org/x/net/context"
)

var httpLog *logging.Logger

// MaxBytesRead limits the number of bytes to read per request
const MaxBytesRead = 10000000

// InvokeRequest describes the body of a structured invoke request
type InvokeRequest struct {
	ID       string   `json:"id"`
	Function string   `json:"function"`
	Params   []string `json:"params"`
}

// InvokeError describes an invoke error
type InvokeError struct {
	Code  string `json:"code"`
	Msg   string `json:"msg"`
	Error bool   `json:"error"`
}

// Success describes a successful response from a cocoon code
type Success struct {
	Body []byte `json:"body"`
}

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

// headerToMap converts http.Header to a map[string]string.
// Header keys with multiple values are joined to as a single string with
// a comma `,` as a delimeter.
func headerToMap(header http.Header) map[string]string {
	var newMap = make(map[string]string)
	for h, v := range header {
		newMap[h] = strings.Join(v, ",")
	}
	return newMap
}

// headerToMap converts url.Values to a map[string]string.
// Values keys with multiple values are joined to as a single string with
// a comma `,` as a delimeter.
func valuesToMap(values url.Values) map[string]string {
	var newMap = make(map[string]string)
	for h, v := range values {
		newMap[h] = strings.Join(v, ",")
	}
	return newMap
}

// prepareHeader collects and sets values to be included in the header
// that will be sent to the cocoon code.
func prepareHeader(w http.ResponseWriter, r *http.Request) http.Header {
	header := r.Header
	header.Set("Method", r.Method)
	header.Set("Host", r.Host)
	header.Set("Remote-Addr", r.RemoteAddr)
	if len(r.Cookies()) > 0 {
		cookiesToJSON, _ := util.ToJSON(r.Cookies())
		header.Set("Cookies", string(cookiesToJSON))
	}
	header.Set("Referer", r.Referer())
	header.Set("URL", r.URL.String())
	header.Set("Request-Uri", r.RequestURI)
	if len(r.URL.Query()) > 0 {
		queryToJSON, _ := util.ToJSON(valuesToMap(r.URL.Query()))
		header.Set("Query", string(queryToJSON))
	}

	r.ParseForm()
	r.ParseMultipartForm(MaxBytesRead)
	if len(r.PostForm) > 0 {
		formValToJSON, _ := util.ToJSON(valuesToMap(r.PostForm))
		header.Set("Form", string(formValToJSON))
	}
	return header
}

func (s *HTTP) invokeCocoonCode(w http.ResponseWriter, r *http.Request) {

	var resp *proto_connector.Response

	// prepare header
	r.Body = http.MaxBytesReader(w, r.Body, MaxBytesRead)
	preparedHeader := prepareHeader(w, r)
	w.Header().Set("Content-Type", "application/json")

	ctx, cc := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cc()

	// attempt to decode body to json
	decoder := json.NewDecoder(r.Body)
	var invokeRequest InvokeRequest
	err := decoder.Decode(&invokeRequest)

	// not json? consider as unstructured only if content type is not application/json
	if err != nil {
		if r.Header.Get("Content-Type") != "application/json" {
			goto unstructured
		}
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(InvokeError{
			Error: true,
			Code:  "1",
			Msg:   "Invalid invoke body structure",
		})
		return
	}
	defer r.Body.Close()

	// no function, consider as unstructured
	if len(strings.TrimSpace(invokeRequest.Function)) == 0 {
		goto unstructured
	}

	// expect ID, if set to be a UUIDv4 string
	if len(invokeRequest.ID) > 0 && !govalidator.IsUUIDv4(invokeRequest.ID) {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(InvokeError{
			Error: true,
			Code:  "2",
			Msg:   "Invalid ID. ID must be a UUIDv4 value",
		})
		return
	}

	// set ID of not set
	if len(invokeRequest.ID) == 0 {
		invokeRequest.ID = util.UUID4()
	}

	preparedHeader.Set("Transaction-Id", invokeRequest.ID)
	preparedHeader.Set("Structured", "yes")

	resp, err = s.rpc.cocoonCodeOps.Handle(ctx, &proto_connector.CocoonCodeOperation{
		ID:       invokeRequest.ID,
		Function: invokeRequest.Function,
		Params:   invokeRequest.Params,
		Header:   headerToMap(preparedHeader),
	})
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(InvokeError{
			Error: true,
			Code:  "code_error",
			Msg:   err.Error(),
		})
		return
	}

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(Success{
		Body: resp.GetBody(),
	})
	return

unstructured:
	preparedHeader.Set("Structured", "no")

	resp, err = s.rpc.cocoonCodeOps.Handle(ctx, &proto_connector.CocoonCodeOperation{
		Header: headerToMap(preparedHeader),
	})
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(InvokeError{
			Error: true,
			Code:  "code_error",
			Msg:   err.Error(),
		})
		return
	}

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(Success{
		Body: resp.GetBody(),
	})
}
