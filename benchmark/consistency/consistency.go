package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/drbg"
	"github.com/privacylab/talek/libtalek"
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
	spotChecks := pflag.Int("spotcheck", 8, "How many cells to check for consistency")
	pflag.Parse()

	// Config
	config := libtalek.ClientConfigFromFile(*configPath)
	if config == nil {
		fmt.Fprintln(os.Stderr, "Talek Client must be run with --config specifying where the server is.")
		os.Exit(1)
	}
	if config.Config == nil {
		fmt.Fprintln(os.Stderr, "Common configuration will be fetched from frontend.")
	}

	leaderRPC := common.NewFrontendRPC("RPC", config.FrontendAddr)

	if config.Config == nil {
		config.Config = new(common.Config)
		err := leaderRPC.GetConfig(nil, config.Config)
		if err != nil {
			panic(err)
		}
	}

	// 1. read through database and make sure replicas are in sync.
	for cell := uint64(0); cell < config.Config.NumBuckets; cell += config.Config.NumBuckets / uint64(*spotChecks) {
		var expectedForCell []byte
		for replica := 0; replica < len(config.TrustDomains); replica++ {
			args := initReadArg(config.Config.NumBuckets, len(config.TrustDomains))
			args.TD[replica].RequestVector[cell/8] ^= (1 << (cell % 8))

			encArgs, _ := args.Encode(config.TrustDomains)
			reply := common.ReadReply{}
			leaderRPC.Read(&encArgs, &reply)
			if replica == 0 {
				expectedForCell = reply.Data
			} else {
				if !bytes.Equal(reply.Data, expectedForCell) {
					fmt.Fprintf(os.Stderr, "Disagreement of value of cell %d between 0 and %d.\n", cell, replica)
				}
			}
		}
		fmt.Fprintf(os.Stderr, ".")
	}
	fmt.Fprintf(os.Stderr, "\n")

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
