package server

import (
	"log"
	"os"

	"github.com/privacylab/talek/common"
)

// FrontendServer is an RPC server for a Frontend API
type FrontendServer struct {
	log     *log.Logger
	name    string
	rpcPort int

	frontend *Frontend
	netRPC   *NetworkRPC
	netHTTP  *NetworkHTTP
}

// NewFrontendServer creates a new FrontEndServer for a given configuration.
func NewFrontendServer(name string, rpcPort int, serverConfig *Config, replicas []common.TrustDomainConfig) *FrontendServer {
	fe := &FrontendServer{}
	fe.log = log.New(os.Stdout, "[FrontendServer:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	fe.name = name
	fe.rpcPort = rpcPort

	rpcs := make([]common.ReplicaInterface, len(replicas))
	for i, r := range replicas {
		rpcs[i] = common.NewReplicaRPC(r.Name, &r)
	}

	fe.frontend = NewFrontend(name, serverConfig, rpcs)
	if rpcPort != 0 {
		fe.netRPC = NewNetworkRPC(fe.frontend, rpcPort)
	}
	//fe.netHttp = NewNetworkHttp(httpPort)
	return fe
}

// Kill stops a frontend server.
func (fe *FrontendServer) Kill() {
	fe.netRPC.Kill()
}
