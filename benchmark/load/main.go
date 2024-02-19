package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/libtalek"
	"github.com/privacylab/talek/server"
)

var numClients = flag.Int("clients", 1, "Number of clients")
var leaderPIR = flag.String("leader", "cpu.0", "PIR backing for leader shard")
var followerPIR = flag.String("follower", "cpu.1", "PIR backing for follower shard")

type killable interface {
	Close() error
}

func main() {
	var err error

	log.Println("Simple Sanity Test")
	s := make(map[string]killable)
	flag.Parse()

	// For trace debug status
	go http.ListenAndServe("localhost:8080", nil)

	// Config
	trustDomainConfig0 := common.NewTrustDomainConfig("t0", "localhost:9000", true, false)
	trustDomainConfig1 := common.NewTrustDomainConfig("t1", "localhost:9001", true, false)
	trustDomainFE := common.NewTrustDomainConfig("f0", "localhost:8999", true, false)
	config := common.ConfigFromFile("../commonconfig.json")
	serverConfig1 := server.ConfigFromFile("../serverconfig.json", config)
	serverConfig1.TrustDomainIndex = 1
	serverConfig1.TrustDomain = trustDomainConfig1
	serverConfig0 := server.ConfigFromFile("../serverconfig.json", config)
	serverConfig0.TrustDomain = trustDomainConfig0
	serverConfigF := server.ConfigFromFile("../serverconfig.json", config)
	serverConfigF.TrustDomain = trustDomainFE

	// Trust Domain 1
	t1 := server.NewReplicaServer("t1", *followerPIR, *serverConfig1)
	s["t1"], err = t1.Run("localhost:9001")
	if err != nil {
		panic(err)
	}

	// Trust Domain 0
	//t0 := server.NewCentralized("t0", config, t1, true)
	t0 := server.NewReplicaServer("t0", *leaderPIR, *serverConfig0)
	s["t0"], _ = t0.Run("localhost:9000")

	// Frontend
	f0 := server.NewFrontendServer("f0", serverConfigF, []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1})
	s["f0"], _ = f0.Run("localhost:8999")

	// Client
	clientConfig := libtalek.ClientConfig{
		Config:        config,
		WriteInterval: time.Second,
		ReadInterval:  time.Second,
		TrustDomains:  []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1},
	}
	//c0 := libtalek.NewClient("c0", config, t0)
	//c1 := libtalek.NewClient("c1", config, t0)
	clients := make([]*libtalek.Client, *numClients)
	clientLeaderSock := common.NewFrontendRPC("c0->f0", trustDomainFE.Address)
	for i := 0; i < *numClients; i++ {
		clients[i] = libtalek.NewClient("c"+fmt.Sprintf("%d", i), clientConfig, clientLeaderSock)
		handle, _ := libtalek.NewTopic()
		var origHandle libtalek.Handle
		origHandle = handle.Handle
		clients[i].Publish(handle, []byte("Hello from client"+fmt.Sprintf("%d", i)))
		fmt.Printf("Published. Waiting for response.")
		data := clients[i].Poll(&origHandle)
		_ = <-data
		fmt.Printf("Client Roundtrip.")
	}
	//c1.Ping()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	for _, v := range s {
		v.Close()
	}
}
