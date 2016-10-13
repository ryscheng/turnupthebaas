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
	emptyTrustDomainConfig := common.NewTrustDomainConfig("", "", false, false)
	globalConfig := common.GlobalConfig{100, 1024, time.Second, time.Second, []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1}}

	// Trust Domain 1
	s["t1"] = server.NewCentralizedFollowerServer("t1", 9001, emptyTrustDomainConfig, false)

	// Trust Domain 0
	s["t0"] = server.NewCentralizedLeaderServer("t0", 9000, trustDomainConfig1, true)

	// Client
	c := libpdb.NewClient("c0", globalConfig)
	c.Ping()
	time.Sleep(10 * time.Second)

	// Kill servers
	for _, v := range s {
		v.Kill()
	}
}
