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

	// Trust Domain Config
	trustDomainConfig1 := common.NewTrustDomainConfig("t1", "localhost:9000", true)
	trustDomainConfig2 := common.NewTrustDomainConfig("t2", "localhost:9010", true)
	emptyTrustDomainConfig := common.NewTrustDomainConfig("", "", false)

	// Trust Domain 2
	dataLayerConfig2 := &server.DataLayerConfig{map[string]map[string]string{"t2g1": map[string]string{"t2g1s1": "localhost:9011"}}}
	s["t2g1s1"] = server.NewShardServer("t2g1", "t2g1s1", 9011, dataLayerConfig2)
	s["t2fe1"] = server.NewFrontendServer("t2fe1", 9010, dataLayerConfig2, emptyTrustDomainConfig, false)

	// Trust Domain 1
	dataLayerConfig1 := &server.DataLayerConfig{map[string]map[string]string{"t1g1": map[string]string{"t1g1s1": "localhost:9001"}}}
	s["t1g1s1"] = server.NewShardServer("t1g1", "t1g1s1", 9001, dataLayerConfig1)
	s["t1fe1"] = server.NewFrontendServer("t1fe1", 9000, dataLayerConfig1, trustDomainConfig2, true)

	// Client
	c := libpdb.NewClient("c1", trustDomainConfig1)
	c.Ping()
	time.Sleep(10 * time.Second)

	// Kill servers
	for _, v := range s {
		v.Kill()
	}
}
