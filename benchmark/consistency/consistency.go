package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/drbg"
	"github.com/privacylab/talek/libtalek"
	"github.com/privacylab/talek/server"
	"github.com/spf13/pflag"
)

func initReadArg(buckets uint64, replicas int) *common.ReadArgs {
	args := common.ReadArgs{}
	args.TD = make([]common.PirArgs, replicas)
	for i := 0; i < replicas; i++ {
		args.TD[i].PadSeed = make([]byte, drbg.SeedLength)
		args.TD[i].RequestVector = make([]byte, buckets/8)
	}
	return &args
}

// Consistency acts as a client driver against a talek system to verify that
// consistency guarantees are enforced. It will write into a cell, and then
// perform a set of reads to ensure that all servers expose the same snapshot
// of the database and are synchronized.
func main() {
	configPath := pflag.String("config", "talek.conf", "Client configuration for talek")
	integrated := pflag.Bool("integrated", false, "Local benchmark in a single process")
	spotChecks := pflag.Int("spotcheck", 8, "How many cells to check for consistency")
	pflag.Parse()

	var config *libtalek.ClientConfig
	if *integrated {
		//Common Config
		conf := &common.Config{
			NumBuckets:         uint64(1024),
			BucketDepth:        uint64(4),
			DataSize:           uint64(256),
			BloomFalsePositive: float64(0.05),
			WriteInterval:      time.Second * 5,
			ReadInterval:       time.Second * 5,
			MaxLoadFactor:      float64(0.95),
			LoadFactorStep:     float64(0.05),
		}
		//Trust domains
		td1 := common.NewTrustDomainConfig("td1", "localhost:9001", true, false)
		td2 := common.NewTrustDomainConfig("td2", "localhost:9002", true, false)
		sc1 := server.Config{Config: conf, WriteInterval: time.Second, ReadBatch: 4, TrustDomain: td1}
		sc2 := server.Config{Config: conf, ReadBatch: 4, TrustDomain: td2, TrustDomainIndex: 1}
		//replicas
		r1 := server.NewCentralized("r1", "cpu.0", sc1)
		r2 := server.NewCentralized("r2", "cpu.0", sc2)
		//client
		config = &libtalek.ClientConfig{Config: conf, WriteInterval: time.Second, ReadInterval: time.Second, TrustDomains: []*common.TrustDomainConfig{td1, td2}, FrontendAddr: "localhost:9000"}
		//frontend
		f0 := server.NewFrontend("f0", &sc1, []common.ReplicaInterface{common.ReplicaInterface(r1), common.ReplicaInterface(r2)})
		f0.Verbose = true
		server.NewNetworkRPC(f0, 9000)
	} else {
		// Config
		config = libtalek.ClientConfigFromFile(*configPath)
		if config == nil {
			fmt.Fprintln(os.Stderr, "Talek Client must be run with --config specifying where the server is.")
			os.Exit(1)
		}
		if config.Config == nil {
			fmt.Fprintln(os.Stderr, "Common configuration will be fetched from frontend.")
		}
	}

	leaderRPC := common.NewFrontendRPC("RPC", config.FrontendAddr)

	if config.Config == nil {
		config.Config = new(common.Config)
		err := leaderRPC.GetConfig(nil, config.Config)
		if err != nil {
			panic(err)
		}
	}

	if !WritesPersisted(*config, *leaderRPC, *spotChecks) {
		return
	}

	c0 := libtalek.NewClient("testClient", *config, leaderRPC)
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
