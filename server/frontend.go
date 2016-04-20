package server

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Frontend struct {
	log        *log.Logger
	httpServer *http.Server
	mux        *http.ServeMux
}

func New(port int) *Frontend {
	fe := &Frontend{}
	fe.log = log.New(os.Stdout, "[server] ", log.Ldate|log.Ltime|log.Lshortfile)
	fe.mux = http.NewServeMux()
	fe.mux.HandleFunc("/ping", fe.handlePing)
	fe.mux.HandleFunc("/", fe.handleDefault)
	fe.httpServer = &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: fe.mux,
	}
	fe.log.Println("New: starting new server on port " + strconv.Itoa(port))
	log.Fatal(fe.httpServer.ListenAndServe())
	return fe
}

func (fe *Frontend) handleDefault(w http.ResponseWriter, r *http.Request) {
	fe.log.Println("handleDefault: " + r.Method + " from " + r.RemoteAddr + ", ... dropping")
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.NotFound(w, r)
}

func (fe *Frontend) handlePing(w http.ResponseWriter, r *http.Request) {
	fe.log.Println("handlePing: " + r.Method + " from " + r.RemoteAddr + ", ... PONG")
	io.WriteString(w, "PONG")
}
