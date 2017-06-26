package server

import (
	"encoding/json"
	"net/http"
	"time"

	"strings"

	"net/url"

	"io/ioutil"

	"net"

	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/ellcrys/cocoon/core/config"
	"github.com/ellcrys/cocoon/core/connector/server/proto_connector"
	"github.com/ellcrys/util"
	"github.com/gorilla/mux"
	logging "github.com/op/go-logging"
	"github.com/pkg/errors"
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

// prepareInvokeHeader collects and sets values to be included in the header
// that will be sent to the cocoon code.
func prepareInvokeHeader(w http.ResponseWriter, r *http.Request) http.Header {
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

// HTTP defines a structure for an HTTP server
// that provides REST API services.
type HTTP struct {
	rpc *RPC
}

// NewHTTP creates an new http server instance
func NewHTTP(rpc *RPC) *HTTP {
	httpLog = config.MakeLogger("connector.http")
	return &HTTP{rpc}
}

// getRouter returns the router
func (s *HTTP) getRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", s.index)

	// v1 API routes
	v1 := r.PathPrefix("/v1").Subrouter()
	v1.HandleFunc("/invoke", s.invokeCocoonCode)

	return r
}

// Start starts the http server. Passes true to the startedCh channel
// when started
func (s *HTTP) Start(addr string, startedCh chan bool) error {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		log.Errorf("%+v", errors.Wrap(err, "failed to parse address"))
		startedCh <- false
		return err
	}

	time.AfterFunc(2*time.Second, func() {
		httpLog.Infof("Started HTTP API server @ :%s", port)
		startedCh <- true
	})

	return http.ListenAndServe(addr, s.getRouter())
}

// index page
func (s *HTTP) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	fmt.Fprint(w, "Hello!")
}

// invokeCocoonCode handles cocoon code invocation
func (s *HTTP) invokeCocoonCode(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

	var err error
	var resp *proto_connector.Response
	var invokeRequest InvokeRequest
	var body []byte
	var ctx context.Context
	var cc context.CancelFunc

	defer r.Body.Close()

	// prepare header
	r.Body = http.MaxBytesReader(w, r.Body, MaxBytesRead)
	preparedHeader := prepareInvokeHeader(w, r)

	// set response content type
	w.Header().Set("Content-Type", "application/json")

	// get request are considered as unstructured
	if r.Method == "GET" {
		goto UNSTRUCTURED
	}

	// get body
	if r.Method == "POST" {
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(InvokeError{
				Error: true,
				Code:  "server_error",
				Msg:   "Invalid invoke body structure",
			})
			return
		}
	}

	// non-json content type are considered as unstructured request
	if r.Header.Get("Content-Type") != "application/json" {
		goto UNSTRUCTURED
	}

	// attempt to decode json body to an invoke request structure
	err = util.FromJSON(body, &invokeRequest)
	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(InvokeError{
			Error: true,
			Code:  "1",
			Msg:   "Invalid invoke body structure",
		})
		return
	}

	// expect ID, if set to be a UUIDv4 string
	if len(invokeRequest.ID) > 0 && !govalidator.IsUUIDv4(invokeRequest.ID) {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(InvokeError{
			Error: true,
			Code:  "2",
			Msg:   "invalid ID. ID must be a UUIDv4 value",
		})
		return
	}

	// return error if request function is not set
	if len(strings.TrimSpace(invokeRequest.Function)) == 0 {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(InvokeError{
			Error: true,
			Code:  "3",
			Msg:   "function is required",
		})
		return
	}

	// set ID of not set
	if len(invokeRequest.ID) == 0 {
		invokeRequest.ID = util.UUID4()
	}

	preparedHeader.Set("Transaction-Id", invokeRequest.ID)
	preparedHeader.Set("Structured", "yes")

	ctx, cc = context.WithTimeout(context.Background(), 2*time.Minute)
	defer cc()
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

	if r.Method == "OPTIONS" {
		return
	}

	json.NewEncoder(w).Encode(Success{
		Body: resp.GetBody(),
	})
	return

UNSTRUCTURED:

	preparedHeader.Set("Structured", "no")

	if r.Method != "GET" {
		preparedHeader.Set("Raw-Body", string(body))
	}

	ctx, cc = context.WithTimeout(context.Background(), 2*time.Minute)
	defer cc()
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
