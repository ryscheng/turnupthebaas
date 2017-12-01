package server

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
)

// ReplicaServer is an RPC server for a Replica
type ReplicaServer struct {
	log  *log.Logger
	name string

	Replica *Replica
	*rpc.Server
}

// NewReplicaServer creates a new Replica served over HTTP
func NewReplicaServer(name string, backing string, serverConfig Config) *ReplicaServer {
	r := &ReplicaServer{}
	r.log = log.New(os.Stdout, "[ReplicaServer:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	r.name = name

	r.Replica = NewReplica(name, backing, serverConfig)

	// Set up the RPC server component.
	r.Server = rpc.NewServer()
	r.Server.RegisterCodec(json.NewCodec(), "application/json")
	r.Server.RegisterTCPService(r.Replica, "Replica")

	return r
}

// Run begins an HTTP server for the server at a specific address
func (r *ReplicaServer) Run(address string) (net.Listener, error) {
	if r.Server == nil {
		r.Server = rpc.NewServer()
		r.Server.RegisterCodec(&json.Codec{}, "application/json")
		r.Server.RegisterTCPService(r.Replica, "Replica")
	}

	bindAddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP("tcp4", bindAddr)
	if err != nil {
		return nil, err
	}
	go http.Serve(listener, r)

	return listener, nil
}
