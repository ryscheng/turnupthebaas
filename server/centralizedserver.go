package server

import (
	"github.com/ryscheng/pdb/common"
	"log"
	"os"
)

type CentralizedServer struct {
	log     *log.Logger
	name    string
	rpcPort int

	centralized *Centralized
	netRpc      *NetworkRpc
	netHttp     *NetworkHttp
}

func NewCentralizedServer(name string, rpcPort int, follower *common.TrustDomainConfig, isLeader bool) *CentralizedServer {
	cs := &CentralizedServer{}
	cs.log = log.New(os.Stdout, "[CentralizedServer:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	cs.name = name
	cs.rpcPort = rpcPort

	cs.centralized = NewCentralized(name, follower, isLeader)
	if rpcPort != 0 {
		cs.netRpc = NewNetworkRpc(cs.centralized, rpcPort)
	}
	return cs
}

func (cs *CentralizedServer) Kill() {
	cs.netRpc.Kill()
}
