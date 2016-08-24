package server

import (
	"github.com/ryscheng/pdb/common"
	"log"
	"os"
)

type FrontendServer struct {
	log     *log.Logger
	name    string
	rpcPort int

	frontend *Frontend
	netRpc   *NetworkRpc
	netHttp  *NetworkHttp
}

func NewFrontendServer(name string, rpcPort int, dataLayerConfig *DataLayerConfig, follower *common.TrustDomainConfig, isLeader bool) *FrontendServer {
	fe := &FrontendServer{}
	fe.log = log.New(os.Stdout, "[FrontendServer:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	fe.name = name
	fe.rpcPort = rpcPort

	fe.frontend = NewFrontend(name, dataLayerConfig, follower, isLeader)
	if rpcPort != 0 {
		fe.netRpc = NewNetworkRpc(fe.frontend, rpcPort)
	}
	//fe.netHttp = NewNetworkHttp(httpPort)
	return fe
}

func (fe *FrontendServer) Kill() {
	fe.netRpc.Kill()
}
