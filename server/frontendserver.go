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
func NewFrontendServer(name string, rpcPort int, serverConfig *Config, follower *common.TrustDomainConfig, isLeader bool) *FrontendServer {
	fe := &FrontendServer{}
	fe.log = log.New(os.Stdout, "[FrontendServer:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	fe.name = name
	fe.rpcPort = rpcPort

	fe.frontend = NewFrontend(name, serverConfig, nil)
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
