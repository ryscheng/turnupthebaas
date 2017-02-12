package main

import (
	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/libtalek"
	"log"
	"math/rand"
	"sort"
	"time"
)

type ByTime []time.Duration

func (d ByTime) Len() int           { return len(d) }
func (d ByTime) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d ByTime) Less(i, j int) bool { return d[i].Nanoseconds() < d[j].Nanoseconds() }

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
	config.TrustDomains = []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1, trustDomainConfig2}

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
	c0 := libtalek.NewClient("c0", *config, leaderRpc)
	log.Println("Created c0")
	time.Sleep(time.Duration(rand.Int()%int(config.WriteInterval)) * time.Nanosecond)
	c1 := libtalek.NewClient("c1", *config, leaderRpc)
	log.Println("Created c1")

	totalTrials := 10
	durations := make([]time.Duration, 0)
	for trials := 0; trials < totalTrials; trials++ {
		time.Sleep(time.Duration(rand.Int()%int(config.WriteInterval)) * time.Nanosecond)
		startTime := time.Now()
		seqNo := c0.PublishTrace()
		log.Printf("c0.Publish -> seqNo=%v after %v\n", seqNo, time.Since(startTime))
		for i := 0; true; i++ {
			seqNoRange := c1.PollTrace()
			log.Printf("c1.Poll#%v: range=%v after %v\n", i, seqNoRange, time.Since(startTime))
			if seqNoRange.Contains(seqNo) {
				log.Printf("Trial#%v: seqNo=%v in range=%v after %v\n", trials, seqNo, seqNoRange, time.Since(startTime))
				durations = append(durations, time.Since(startTime))
				break
			}
		}
	}
	log.Printf("Done\n")
	sort.Sort(ByTime(durations))
	log.Printf("min=%v, q1=%v, median=%v, q3=%v, max=%v\n", durations[0], durations[totalTrials/4], durations[totalTrials/2], durations[3*totalTrials/4], durations[totalTrials-1])

	/**
	// Go on forever
	for {
		time.Sleep(10 * time.Second)
	}
	**/
}
