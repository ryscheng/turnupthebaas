package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/libtalek"
	"github.com/privacylab/talek/server"
)

var configPath = flag.String("config", "talek.conf", "Talek Client Configuration")
var systemPath = flag.String("server", "server.conf", "Talek Server Configuration")

// Starts a talek frontend operating with configuration from talekutil
func main() {
	log.Println("----------------------")
	log.Println("--- Talek Frontend ---")
	log.Println("----------------------")
	flag.Parse()

	config := libtalek.ClientConfigFromFile(*configPath)
	serverConfig := server.ConfigFromFile(*systemPath, config.Config)

	replicas := make([]common.ReplicaInterface, len(config.TrustDomains))
	for i, td := range config.TrustDomains {
		replicas[i] = common.NewReplicaRPC(td.Name, td)
	}

	f := server.NewFrontend("Talek Frontend", serverConfig, replicas)
	_, port, _ := net.SplitHostPort(config.FrontendAddr)
	pnum, _ := strconv.Atoi(port)
	_ = server.NewNetworkRPC(f, pnum)

	log.Println("Running.")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	f.Close()
}
