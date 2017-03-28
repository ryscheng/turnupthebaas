package main

import (
	"log"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/libtalek"
	"github.com/privacylab/talek/server"
)

type Killable interface {
	Kill()
}

func main() {
	log.Println("Simple Sanity Test")
	s := make(map[string]Killable)

	// Config
	trustDomainConfig0 := common.NewTrustDomainConfig("t0", "localhost:9000", true, true)
	trustDomainConfig1 := common.NewTrustDomainConfig("t1", "localhost:9100", true, true)
	emptyTrustDomainConfig := common.NewTrustDomainConfig("", "", false, true)
	config := common.CommonConfigFromFile("commonconfig.json")
	serverConfig0 := server.ConfigFromFile("serverconfig.json", config)
	serverConfig0.TrustDomain = trustDomainConfig0
	serverConfig1 := server.ConfigFromFile("serverconfig.json", config)
	serverConfig1.TrustDomain = trustDomainConfig1
	serverConfig1.TrustDomainIndex = 1

	// Trust Domain 1
	serverConfig1.ServerAddrs = map[string]map[string]string{
		"t1g0": map[string]string{
			"t1g0s0": "localhost:9101",
		},
	}
	shard1 := server.NewShard("t1g0s0", "pir.socket", *serverConfig1)
	s["t1g0s0"] = server.NewNetworkRpc(shard1, 9101)
	s["t1fe0"] = server.NewFrontendServer("t1fe0", 9100, serverConfig1, emptyTrustDomainConfig, false)

	// Trust Domain 0
	serverConfig0.ServerAddrs = map[string]map[string]string{
		"t0g0": map[string]string{
			"t0g0s0": "localhost:9001",
		},
	}
	shard0 := server.NewShard("t0g0s0", "pir2.socket", *serverConfig0)
	s["t0g0s0"] = server.NewNetworkRpc(shard0, 9001)
	s["t0fe0"] = server.NewFrontendServer("t0fe0", 9000, serverConfig0, trustDomainConfig1, true)

	// Client
	clientConfig := libtalek.ClientConfig{
		CommonConfig:  config,
		WriteInterval: time.Second,
		ReadInterval:  time.Second,
		TrustDomains:  []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1},
	}
	clientLeaderSock := common.NewLeaderRpc("c0->t0", trustDomainConfig1)
	c := libtalek.NewClient("c1", clientConfig, clientLeaderSock)
	c.Ping()
	time.Sleep(10 * time.Second)

	// Kill servers
	for _, v := range s {
		v.Kill()
	}
}
