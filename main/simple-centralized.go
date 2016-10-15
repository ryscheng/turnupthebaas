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
	trustDomainConfig0 := common.NewTrustDomainConfig("t0", "localhost:9000", true, false)
	trustDomainConfig1 := common.NewTrustDomainConfig("t1", "localhost:9100", true, false)
	globalConfig := common.GlobalConfig{100, 2, 100, 1024, time.Second, time.Second, []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1}}

	// Trust Domain 1
	t1 := server.NewCentralized("t1", globalConfig, nil, false)
	//s["t1"] = NewNetworkRpc(t1, 9001)

	// Trust Domain 0
	t0 := server.NewCentralized("t0", globalConfig, t1, true)
	//t0 := server.NewCentralized("t0", common.NewLeaderRpc("t0->t1", trustDomainConfig1), true)
	//s["t0"] = NewNetworkRpc(t0, 9000)

	// Client
	c0 := libpdb.NewClient("c0", globalConfig, t0)
	c1 := libpdb.NewClient("c1", globalConfig, t0)
	//c := libpdb.NewClient("c0", globalConfig, common.NewLeaderRpc("c0->t0", trustDomainConfig0))
	c0.Ping()
	c1.Ping()
	time.Sleep(10 * time.Second)

	// Kill servers
	for _, v := range s {
		v.Kill()
	}
}
