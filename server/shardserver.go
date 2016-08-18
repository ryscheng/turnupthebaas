package server

import (
	"log"
	"os"
)

type ShardServer struct {
	log     *log.Logger
	shard   *Shard
	netRpc  *NetworkRpc
	netHttp *NetworkHttp
}

func NewShardServer(rpcPort int, httpPort int) *ShardServer {
	s := &ShardServer{}
	s.log = log.New(os.Stdout, "[ShardServer] ", log.Ldate|log.Ltime|log.Lshortfile)
	s.shard = NewShard()
	if rpcPort != 0 {
		s.netRpc = NewNetworkRpc(s.shard, rpcPort)
	}
	if httpPort != 0 {
		s.netHttp = NewNetworkHttp(httpPort)
	}
	return s
}

func (s *ShardServer) Kill() {
	s.netRpc.Kill()
}
