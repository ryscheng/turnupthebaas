package server

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

type FrontEndHttp struct {
	log        *log.Logger
	port       int
	mux        *http.ServeMux
	httpServer *http.Server
}

func NewFrontEndHttp(port int) *FrontEndHttp {
	fe := &FrontEndHttp{}
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
	fe.log.Println("NewFrontEndHttp: starting new server on port " + strconv.Itoa(port))
	go fe.log.Fatal(fe.httpServer.ListenAndServe())
	return fe
}

func (fe *FrontEndHttp) handleDefault(w http.ResponseWriter, r *http.Request) {
	fe.log.Println("handleDefault: " + r.Method + " from " + r.RemoteAddr + ", ... dropping")
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.NotFound(w, r)
}

func (fe *FrontEndHttp) handlePing(w http.ResponseWriter, r *http.Request) {
	fe.log.Println("handlePing: " + r.Method + " from " + r.RemoteAddr + ", ... PONG")
	io.WriteString(w, "PONG")
}
