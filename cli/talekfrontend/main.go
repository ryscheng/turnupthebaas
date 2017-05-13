package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"

	"github.com/coreos/etcd/pkg/flags"
	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/libtalek"
	"github.com/privacylab/talek/server"
	"github.com/spf13/pflag"
)

// Starts a talek frontend operating with configuration from talekutil
func main() {
	log.Println("----------------------")
	log.Println("--- Talek Frontend ---")
	log.Println("----------------------")

	configPath := pflag.String("client", "talek.conf", "Talek Client Configuration")
	commonPath := pflag.String("common", "common.conf", "Talek Common Configuration")
	systemPath := pflag.String("server", "server.conf", "Talek Server Configuration")
	err := flags.SetPflagsFromEnv(common.EnvPrefix, pflag.CommandLine)
	if err != nil {
		log.Printf("Error reading environment variables, %v\n", err)
		return
	}
	pflag.Parse()

	config := libtalek.ClientConfigFromFile(*configPath)
	config.Config = common.ConfigFromFile(*commonPath)
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
