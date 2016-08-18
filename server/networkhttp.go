package server

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

type NetworkHttp struct {
	log        *log.Logger
	port       int
	mux        *http.ServeMux
	httpServer *http.Server
}

func NewNetworkHttp(port int) *NetworkHttp {
	fe := &NetworkHttp{}
	fe.log = log.New(os.Stdout, "[frontendhttp] ", log.Ldate|log.Ltime|log.Lshortfile)
	fe.port = port
	// HTTP Handlers
	fe.mux = http.NewServeMux()
	fe.mux.HandleFunc("/ping", fe.handlePing)
	fe.mux.HandleFunc("/", fe.handleDefault)
	// HTTP Server
	fe.httpServer = &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: fe.mux,
	}
	// Start services
	fe.log.Println("NewNetworkHttp: starting new server on port " + strconv.Itoa(port))
	go fe.log.Fatal(fe.httpServer.ListenAndServe())
	return fe
}

func (fe *NetworkHttp) handleDefault(w http.ResponseWriter, r *http.Request) {
	fe.log.Println("handleDefault: " + r.Method + " from " + r.RemoteAddr + ", ... dropping")
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.NotFound(w, r)
}

func (fe *NetworkHttp) handlePing(w http.ResponseWriter, r *http.Request) {
	fe.log.Println("handlePing: " + r.Method + " from " + r.RemoteAddr + ", ... PONG")
	io.WriteString(w, "PONG")
}
