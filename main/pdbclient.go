package main

import (
	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/libtalek"
	"log"
	"math/rand"
	"time"
)

func main() {
	log.Println("------------------")
	log.Println("--- Talek Client ---")
	log.Println("------------------")

	// Config
	//trustDomainConfig0 := common.NewTrustDomainConfig("t0", "localhost:9000", true, false)
	trustDomainConfig0 := common.NewTrustDomainConfig("t0", "172.30.2.10:9000", true, false)
	trustDomainConfig1 := common.NewTrustDomainConfig("t1", "172.30.2.159:9000", true, false)
	trustDomainConfig2 := common.NewTrustDomainConfig("t2", "172.30.2.221:9000", true, false)
	config := common.CommonConfigFromFile("globalconfig.json")

	leaderRpc := common.NewLeaderRpc("c0->t0", trustDomainConfig0)
	/**
	// Throughput
	numClients := 10000
	for i := 0; i < numClients; i++ {
		_ = libtalek.NewClient("c", *config, leaderRpc)
		time.Sleep(time.Duration(rand.Int()%(2*int(config.WriteInterval)/numClients)) * time.Nanosecond)
	}
	log.Printf("Generated %v clients\n", numClients)
	**/
	//c.Ping()

	// Latency
	clientConfig := libtalek.ClientConfig{
		CommonConfig:  config,
		ReadInterval:  time.Second,
		WriteInterval: time.Second,
		TrustDomains:  []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1, trustDomainConfig2},
	}
	c0 := libtalek.NewClient("c0", clientConfig, leaderRpc)
	log.Println("Created c0")
	time.Sleep(time.Duration(rand.Int()%int(clientConfig.WriteInterval)) * time.Nanosecond)
	c1 := libtalek.NewClient("c1", clientConfig, leaderRpc)
	log.Println("Created c1")

	handle, err := libtalek.NewTopic()
	totalTrials := 10
	durations := make([]time.Duration, 0)
	for trials := 0; trials < totalTrials; trials++ {
		time.Sleep(time.Duration(rand.Int()%int(clientConfig.WriteInterval)) * time.Nanosecond)
		startTime := time.Now()
		err := c0.Publish(handle, []byte("PDB Client Trial"))
		data := c1.Poll(&handle.Subscription)
	}
	log.Printf("Done\n")
}
