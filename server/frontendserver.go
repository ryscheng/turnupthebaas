package server

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/privacylab/talek/common"
)

// FrontendServer represents an HTTP served Frontend
type FrontendServer struct {
	log  *log.Logger
	name string

	Frontend *Frontend
	*rpc.Server
}

// NewFrontendServer creates a new Frontend implementing HTTP.Handler.
func NewFrontendServer(name string, serverConfig *Config, replicas []*common.TrustDomainConfig) *FrontendServer {
	fe := &FrontendServer{}
	fe.log = log.New(os.Stdout, "[FrontendServer:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	fe.name = name

	rpcs := make([]common.ReplicaInterface, len(replicas))
	for i, r := range replicas {
		rpcs[i] = common.NewReplicaRPC(r.Name, r)
	}

	fe.Frontend = NewFrontend(name, serverConfig, rpcs)

	// Set up the RPC server component.
	fe.Server = rpc.NewServer()
	fe.Server.RegisterCodec(&json.Codec{}, "application/json")
	fe.Server.RegisterTCPService(fe.Frontend, "Frontend")

	return fe
}

// Run begins an HTTP server for the server at a specific address
func (fe *FrontendServer) Run(address string) (net.Listener, error) {
	if fe.Server == nil {
		fe.Server = rpc.NewServer()
		fe.Server.RegisterCodec(json.NewCodec(), "application/json")
		fe.Server.RegisterTCPService(fe.Frontend, "Frontend")
	}

	bindAddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP("tcp4", bindAddr)
	if err != nil {
		return nil, err
	}
	go http.Serve(listener, fe)

	return listener, nil
}
