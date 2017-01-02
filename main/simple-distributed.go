package main

import (
	"github.com/ryscheng/pdb/common"
	"github.com/ryscheng/pdb/libpdb"
	"github.com/ryscheng/pdb/server"
	"log"
	"time"
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
	config := common.CommonConfigFromFile("globalconfig.json")
	config.TrustDomains = []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1}

	// Trust Domain 1
	serverConfig1 := &server.ServerConfig{map[string]map[string]string{
		"t1g0": map[string]string{
			"t1g0s0": "localhost:9101",
		},
	}}
	s["t1g0s0"] = server.NewShardServer("t1g0", "t1g0s0", 9101, serverConfig1)
	s["t1fe0"] = server.NewFrontendServer("t1fe0", 9100, serverConfig1, emptyTrustDomainConfig, false)

	// Trust Domain 0
	serverConfig0 := &server.ServerConfig{map[string]map[string]string{
		"t0g0": map[string]string{
			"t0g0s0": "localhost:9001",
		},
	}}
	s["t0g0s0"] = server.NewShardServer("t0g0", "t0g0s0", 9001, serverConfig0)
	s["t0fe0"] = server.NewFrontendServer("t0fe0", 9000, serverConfig0, trustDomainConfig1, true)

	// Client
	c := libpdb.NewClient("c1", *config)
	c.Ping()
	time.Sleep(10 * time.Second)

	// Kill servers
	for _, v := range s {
		v.Kill()
	}
}
