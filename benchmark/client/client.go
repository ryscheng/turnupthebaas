package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/libtalek"
)

var configPath = flag.String("config", "../commonconfig.json", "Talek Common Configuration")
var trustDomainPath = flag.String("trust", "../keys/leaderpublic.json,../keys/followerpublic.json", "Server keys (comma separated)")
var repetitions = flag.Int("repetitions", 10, "How many reads and writes to make")

// This benchmark client will create a client that repeatedly attempts
// to read and write to a given talek configuration.
func main() {
	log.Println("--------------------")
	log.Println("--- Talek Client ---")
	log.Println("--------------------")
	flag.Parse()

	// Config
	config := common.CommonConfigFromFile(*configPath)
	domainPaths := strings.Split(*trustDomainPath, ",")
	trustDomains := make([]*common.TrustDomainConfig, len(domainPaths))
	for i, path := range domainPaths {
		tdString, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("Could not read %s!\n", path)
			return
		}
		trustDomains[i] = new(common.TrustDomainConfig)
		if err := json.Unmarshal(tdString, trustDomains[i]); err != nil {
			log.Printf("Could not parse %s: %v\n", path, err)
			return
		}
	}

	leaderRPC := common.NewLeaderRpc("c0->t0", trustDomains[0])

	clientConfig := libtalek.ClientConfig{
		CommonConfig:  config,
		ReadInterval:  time.Second,
		WriteInterval: time.Second,
		TrustDomains:  trustDomains,
	}
	c0 := libtalek.NewClient("testClient", clientConfig, leaderRPC)
	log.Println("Created client")
	//time.Sleep(time.Duration(rand.Int()%int(clientConfig.WriteInterval)) * time.Nanosecond)
	//c1 := libtalek.NewClient("c1", clientConfig, leaderRpc)
	//log.Println("Created c1")

	handle, _ := libtalek.NewTopic()
	var subscription libtalek.Handle
	subscription = handle.Handle
	totalTrials := *repetitions
	durations := make([]time.Duration, 0, totalTrials)
	read := c0.Poll(&subscription)
	for trials := 0; trials < totalTrials; trials++ {
		time.Sleep(time.Duration(int(clientConfig.WriteInterval*2)) * time.Nanosecond)
		startTime := time.Now()
		// Each 2 rounds the client will publish and make a valid pair of read requests.
		c0.Publish(handle, []byte("PDB Client Trial "+string(trials)))
		<-read
		endTime := time.Now()

		durations = append(durations, endTime.Sub(startTime))
	}
	log.Printf("Done\n")
}
