package main

import (
	"github.com/ryscheng/pdb/common"
	"github.com/ryscheng/pdb/libpdb"
	"log"
	"math/rand"
	"time"
)

func main() {
	log.Println("------------------")
	log.Println("--- PDB Client ---")
	log.Println("------------------")

	// Config
	//trustDomainConfig0 := common.NewTrustDomainConfig("t0", "localhost:9000", true, false)
	//trustDomainConfig1 := common.NewTrustDomainConfig("t1", "localhost:9001", true, false)
	trustDomainConfig0 := common.NewTrustDomainConfig("t0", "172.30.2.10:9000", true, false)
	trustDomainConfig1 := common.NewTrustDomainConfig("t1", "172.30.2.221:9000", true, false)
	globalConfig := common.GlobalConfigFromFile("globalconfig.json")
	globalConfig.TrustDomains = []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1}

	leaderRpc := common.NewLeaderRpc("c0->t0", trustDomainConfig0)
	numClients := 10000
	for i := 0; i < numClients; i++ {
		_ = libpdb.NewClient("c", *globalConfig, leaderRpc)
		time.Sleep(time.Duration(rand.Int()%(2000000/numClients)) * time.Microsecond)
	}
	log.Printf("Generated %v clients\n", numClients)
	//c.Ping()

	for {
		time.Sleep(10 * time.Second)
	}
}
