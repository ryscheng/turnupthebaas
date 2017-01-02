package server

import (
	"log"
	"os"
)

type ShardServer struct {
	log             *log.Logger
	group           string
	name            string
	rpcPort         int
	serverConfig    *ServerConfig

	shard   *Shard
	netRpc  *NetworkRpc
	netHttp *NetworkHttp
}

func NewShardServer(group string, name string, rpcPort int, serverConfig *ServerConfig) *ShardServer {
	s := &ShardServer{}
	s.log = log.New(os.Stdout, "[ShardServer:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	s.group = group
	s.name = name
	s.rpcPort = rpcPort
	s.serverConfig = serverConfig

	s.shard = NewShard(name, "pir.socket", *serverConfig.CommonConfig)
	if rpcPort != 0 {
		s.netRpc = NewNetworkRpc(s.shard, rpcPort)
	}
	//s.netHttp = NewNetworkHttp(httpPort)
	return s
}

func (s *ShardServer) Kill() {
	s.netRpc.Kill()
}
