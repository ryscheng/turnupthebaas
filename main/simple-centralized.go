package main

import (
	"flag"
	"github.com/ryscheng/pdb/common"
	"github.com/ryscheng/pdb/libpdb"
	"github.com/ryscheng/pdb/server"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var numClients = flag.Int("clients", 1, "Number of clients")
var leaderPIR = flag.String("leader", "../pird/pir.socket", "PIR daemon for leader")
var followerPIR = flag.String("follower", "../pird/pir2.socket", "PIR daemon for follower")

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
	globalConfig := common.GlobalConfigFromFile("globalconfig.json")
	globalConfig.TrustDomains = []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1}

	// Trust Domain 1
	t1 := server.NewCentralized("t1", *followerPIR, *globalConfig, nil, false)
	s["t1"] = server.NewNetworkRpc(t1, 9001)

	// Trust Domain 0
	//t0 := server.NewCentralized("t0", globalConfig, t1, true)
	t0 := server.NewCentralized("t0", *leaderPIR, *globalConfig, common.NewFollowerRpc("t0->t1", trustDomainConfig1), true)
	s["t0"] = server.NewNetworkRpc(t0, 9000)

	// Client
	//c0 := libpdb.NewClient("c0", globalConfig, t0)
	//c1 := libpdb.NewClient("c1", globalConfig, t0)
	clients := make([]*libpdb.Client, *numClients)
	for i:=0; i < *numClients; i++ {
		clients[i] = libpdb.NewClient("c" + string(i), *globalConfig, common.NewLeaderRpc("c0->t0", trustDomainConfig0))
		clients[i].Ping()
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
}
