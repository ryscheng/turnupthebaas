package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/libtalek"
	"github.com/privacylab/talek/pir"
	"github.com/privacylab/talek/server"
)

var numClients = flag.Int("clients", 1, "Number of clients")
var leaderPIR = flag.String("leader", "../pird/pir.socket", "PIR daemon for leader")
var followerPIR = flag.String("follower", "../pird/pir2.socket", "PIR daemon for follower")
var mockPIR = flag.Bool("mock", false, "Use the mock PIR daemon")

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
	config := common.ConfigFromFile("commonconfig.json")
	serverConfig1 := server.ConfigFromFile("serverconfig.json", config)
	serverConfig1.TrustDomainIndex = 1
	serverConfig1.TrustDomain = trustDomainConfig1
	serverConfig0 := server.ConfigFromFile("serverconfig.json", config)
	serverConfig0.TrustDomain = trustDomainConfig0

	status := make(chan int)

	if *mockPIR {
		*followerPIR = fmt.Sprintf("pirtest%d.socket", rand.Int())
		*leaderPIR = fmt.Sprintf("pirtest%d.socket", rand.Int())
		go pir.CreateMockServer(status, *followerPIR)
		go pir.CreateMockServer(status, *leaderPIR)
		<-status
		<-status
	}

	// Trust Domain 1
	t1 := server.NewCentralized("t1", *followerPIR, *serverConfig1, nil, false)
	s["t1"] = server.NewNetworkRPC(t1, 9001)

	// Trust Domain 0
	//t0 := server.NewCentralized("t0", config, t1, true)
	t0 := server.NewCentralized("t0", *leaderPIR, *serverConfig0, common.NewFollowerRPC("t0->t1", trustDomainConfig1), true)
	s["t0"] = server.NewNetworkRPC(t0, 9000)

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
	clientLeaderSock := common.NewLeaderRPC("c0->t0", trustDomainConfig0)
	for i := 0; i < *numClients; i++ {
		clients[i] = libtalek.NewClient("c"+string(i), clientConfig, clientLeaderSock)
		clients[i].Ping()
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
	t1.Close()
	t0.Close()
	if *mockPIR {
		status <- 1
		status <- 1
		<-status
		<-status
	}
}
