package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

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
	listen := pflag.StringP("listen", "l", ":8080", "Listening Address")
	verbose := pflag.Bool("verbose", false, "Verbose output")
	err := flags.SetPflagsFromEnv(common.EnvPrefix, pflag.CommandLine)
	if err != nil {
		log.Printf("Error reading environment variables, %v\n", err)
		return
	}
	pflag.Parse()

	config := libtalek.ClientConfigFromFile(*configPath)
	if config == nil {
		pflag.Usage()
		return
	}
	config.Config = common.ConfigFromFile(*commonPath)
	serverConfig := server.ConfigFromFile(*systemPath, config.Config)

	f := server.NewFrontendServer("Talek Frontend", serverConfig, config.TrustDomains)
	f.Frontend.Verbose = *verbose

	bindAddr, err := net.ResolveTCPAddr("ip", *listen)
	if err != nil {
		log.Printf("Couldn't resolve address '%s': %v\n", *listen, err)
		return
	}
	listener, err := net.ListenTCP("ip", bindAddr)
	if err != nil {
		log.Printf("Couldn't listen to frontend address: %v\n", err)
		return
	}
	go http.Serve(listener, f)

	log.Println("Running.")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	listener.Close()
}
