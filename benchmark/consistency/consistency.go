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

// Consistency acts as a client driver against a talek system to verify that
// consistency guarantees are enforced. It will write into a cell, and then
// perform a set of reads to ensure that all servers expose the same snapshot
// of the database and are synchronized.
func main() {
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

	leaderRPC := common.NewLeaderRpc("RPC", trustDomains[0])

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

	topic, _ := libtalek.NewTopic()
	var handle libtalek.Handle
	handle = topic.Handle

	read := c0.Poll(&handle)

	c0.Publish(topic, []byte("PDB Client Trial"))
	log.Printf("waiting for read.")
	data := <-read

	log.Printf("Read back: %v\n", data)
}
