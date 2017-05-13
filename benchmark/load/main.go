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
	Kill()
}

func main() {
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
	t1 := server.NewCentralized("t1", *followerPIR, *serverConfig1)
	s["t1"] = server.NewNetworkRPC(t1, 9001)

	// Trust Domain 0
	//t0 := server.NewCentralized("t0", config, t1, true)
	t0 := server.NewCentralized("t0", *leaderPIR, *serverConfig0)
	s["t0"] = server.NewNetworkRPC(t0, 9000)

	// Frontend
	ft0 := common.NewReplicaRPC("f0-t0", trustDomainConfig0)
	ft1 := common.NewReplicaRPC("f0-t1", trustDomainConfig1)
	f0 := server.NewFrontend("f0", serverConfigF, []common.ReplicaInterface{ft0, ft1})
	server.NewNetworkRPC(f0, 8999)

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
		clients[i] = libtalek.NewClient("c"+string(i), clientConfig, clientLeaderSock)
		handle, _ := libtalek.NewTopic()
		clients[i].Publish(handle, []byte("Hello from client"+string(i)))
		data := clients[i].Poll(&handle.Handle)
		fmt.Printf("!!! data=%v", data)
	}
	//c1.Ping()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	for _, v := range s {
		v.Kill()
	}
	f0.Close()
	t1.Close()
	t0.Close()
}
