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
	servers := make([]Killable, 0)

	dataLayerConfig2 := &server.DataLayerConfig{map[string]map[string]string{"t2g1": map[string]string{"t2g1s1": "localhost:9011"}}}
	t2g1s1 := server.NewShardServer("t2g1", "t2g1s1", 9011, dataLayerConfig2)
	t2fe1 := server.NewFrontendServer("t2fe1", 9010, dataLayerConfig2, &common.TrustDomainConfig{"", false}, false)

	dataLayerConfig1 := &server.DataLayerConfig{map[string]map[string]string{"t1g1": map[string]string{"t1g1s1": "localhost:9001"}}}
	t1g1s1 := server.NewShardServer("t1g1", "t1g1s1", 9001, dataLayerConfig1)
	t1fe1 := server.NewFrontendServer("t1fe1", 9000, dataLayerConfig1, &common.TrustDomainConfig{"localhost:9010", true}, true)

	servers = append(servers, t2g1s1, t2fe1, t1g1s1, t1fe1)

	c := libpdb.NewClient("c1", "localhost:9000")
	c.Ping()

	time.Sleep(10 * time.Second)

	for i := range servers {
		servers[i].Kill()
	}
}
