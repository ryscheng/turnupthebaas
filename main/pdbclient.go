package main

import (
	"github.com/ryscheng/pdb/common"
	"github.com/ryscheng/pdb/libpdb"
	"log"
	"time"
)

func main() {
	log.Println("------------------")
	log.Println("--- PDB Client ---")
	log.Println("------------------")

	// Config
	//trustDomainConfig0 := common.NewTrustDomainConfig("t0", "localhost:9000", true, false)
	//trustDomainConfig1 := common.NewTrustDomainConfig("t1", "localhost:9001", true, false)
	//trustDomainConfig0 := common.NewTrustDomainConfig("t0", "35.163.10.49:9000", true, false)
	trustDomainConfig0 := common.NewTrustDomainConfig("t0", "172.30.2.10:9000", true, false)
	//trustDomainConfig1 := common.NewTrustDomainConfig("t1", "35.163.8.29:9000", true, false)
	trustDomainConfig1 := common.NewTrustDomainConfig("t1", "172.30.2.221:9000", true, false)
	globalConfig := common.GlobalConfigFromFile("globalconfig.json")
	globalConfig.TrustDomains = []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1}

	c := libpdb.NewClient("c", *globalConfig, common.NewLeaderRpc("c0->t0", trustDomainConfig0))
	c.Ping()

	//for {
	time.Sleep(10 * time.Second)
	//}
}
