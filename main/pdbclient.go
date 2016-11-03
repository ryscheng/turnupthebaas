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
	trustDomainConfig0 := common.NewTrustDomainConfig("t0", "localhost:9000", true, false)
	trustDomainConfig1 := common.NewTrustDomainConfig("t1", "localhost:9001", true, false)
	globalConfig := common.GlobalConfigFromFile("globalconfig.json")
	globalConfig.TrustDomains = []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1}


	c := libpdb.NewClient("c", *globalConfig, common.NewLeaderRpc("c0->t0", trustDomainConfig0))
	c.Ping()

  //for {
	  time.Sleep(10 * time.Second)
  //}
}
