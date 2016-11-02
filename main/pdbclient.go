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
	globalConfig := common.GlobalConfig{
		100,         // NumBuckets
		2,           // BucketDepth
		100,         // WindowSize
		1024,        // DataSize
		1,           // ReadBatch
		0.001,       // Bloom FP Rate
		0.90,        // LoadFactor
		0.05,        // LoadFactorStep
		time.Second, // WriteInterval
		time.Second, // ReadInterval
		[]*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1},
	}

	c := libpdb.NewClient("c", globalConfig, common.NewLeaderRpc("c0->t0", trustDomainConfig0))
	c.Ping()
}
