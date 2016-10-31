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
	dataLayerConfig *DataLayerConfig

	shard   *Shard
	netRpc  *NetworkRpc
	netHttp *NetworkHttp
}

func NewShardServer(group string, name string, rpcPort int, dataLayerConfig *DataLayerConfig) *ShardServer {
	s := &ShardServer{}
	s.log = log.New(os.Stdout, "[ShardServer:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	s.group = group
	s.name = name
	s.rpcPort = rpcPort
	s.dataLayerConfig = dataLayerConfig

	s.shard = NewShard(name, *dataLayerConfig.GlobalConfig)
	if rpcPort != 0 {
		s.netRpc = NewNetworkRpc(s.shard, rpcPort)
	}
	//s.netHttp = NewNetworkHttp(httpPort)
	return s
}

func (s *ShardServer) Kill() {
	s.netRpc.Kill()
}
