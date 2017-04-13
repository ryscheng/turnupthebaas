package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/pir"
	"github.com/privacylab/talek/server"
)

var configPath = flag.String("config", "replica.conf", "Talek Replica Configuration")
var pirSocket = flag.String("socket", "../../pird/pir.socket", "PIR daemon socket")

// Starts a single, centralized talek replica operating with configuration from talekutil
func main() {
	log.Println("---------------------")
	log.Println("--- Talek Replica ---")
	log.Println("---------------------")
	flag.Parse()

	configString, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Printf("Could not read %s!\n", *configPath)
		return
	}

	// Default configuration. The server can be started with just a trustdomain
	// config and this will be used for the serverConfig struct in that case.
	serverConfig := server.Config{
		Config:           &common.Config{},
		WriteInterval:    time.Second,
		ReadInterval:     time.Second,
		ReadBatch:        8,
		TrustDomain:      &common.TrustDomainConfig{},
		TrustDomainIndex: 0,
	}
	if err := json.Unmarshal(configString, &serverConfig); err != nil {
		log.Printf("Could not parse %s: %v\n", *configPath, err)
		return
	}

	mockPirStatus := make(chan int)
	usingMock := false
	if len(*pirSocket) == 0 {
		*pirSocket = fmt.Sprintf("pirtest%d.socket", rand.Int())
		go pir.CreateMockServer(mockPirStatus, *pirSocket)
		<-mockPirStatus
		usingMock = true
	}

	s := server.NewCentralized(serverConfig.TrustDomain.Name, *pirSocket, serverConfig)
	_, port, _ := net.SplitHostPort(serverConfig.TrustDomain.Address)
	pnum, _ := strconv.Atoi(port)
	_ = server.NewNetworkRPC(s, pnum)

	log.Println("Running.")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	s.Close()
	if usingMock {
		mockPirStatus <- 1
		<-mockPirStatus
	}
}
