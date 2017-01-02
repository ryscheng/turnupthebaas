package main

import (
	"flag"
	"fmt"
	"github.com/ryscheng/pdb/common"
	"github.com/ryscheng/pdb/libpdb"
	"github.com/ryscheng/pdb/pir"
	"github.com/ryscheng/pdb/server"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
)

var numClients = flag.Int("clients", 1, "Number of clients")
var leaderPIR = flag.String("leader", "../pird/pir.socket", "PIR daemon for leader")
var followerPIR = flag.String("follower", "../pird/pir2.socket", "PIR daemon for follower")
var mockPIR = flag.Bool("mock", false, "Use the mock PIR daemon")

type Killable interface {
	Kill()
}

func main() {
	log.Println("Simple Sanity Test")
	s := make(map[string]Killable)
	flag.Parse()

	// For trace debug status
	go http.ListenAndServe("localhost:8080", nil)

	// Config
	trustDomainConfig0 := common.NewTrustDomainConfig("t0", "localhost:9000", true, false)
	trustDomainConfig1 := common.NewTrustDomainConfig("t1", "localhost:9001", true, false)
	config := common.CommonConfigFromFile("globalconfig.json")
	config.TrustDomains = []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1}

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
	t1 := server.NewCentralized("t1", *followerPIR, *config, nil, false)
	s["t1"] = server.NewNetworkRpc(t1, 9001)

	// Trust Domain 0
	//t0 := server.NewCentralized("t0", config, t1, true)
	t0 := server.NewCentralized("t0", *leaderPIR, *config, common.NewFollowerRpc("t0->t1", trustDomainConfig1), true)
	s["t0"] = server.NewNetworkRpc(t0, 9000)

	// Client
	//c0 := libpdb.NewClient("c0", config, t0)
	//c1 := libpdb.NewClient("c1", config, t0)
	clients := make([]*libpdb.Client, *numClients)
	clientLeaderSock := common.NewLeaderRpc("c0->t0", trustDomainConfig0)
	for i := 0; i < *numClients; i++ {
		clients[i] = libpdb.NewClient("c"+string(i), *config, clientLeaderSock)
		clients[i].Ping()
		seqNo := clients[i].PublishTrace()
		fmt.Printf("!!! seqNo=%v\n", seqNo)
		seqNoRange := clients[i].PollTrace()
		fmt.Printf("!!! seqNo=%v, range=%v\n", seqNo, seqNoRange)
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
