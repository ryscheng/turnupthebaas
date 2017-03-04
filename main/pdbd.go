package main

import (
	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/server"
	"log"
	"net/http"
	"time"
)

func main() {
	log.Println("--------------------")
	log.Println("--- Talek Server ---")
	log.Println("--------------------")

	// For trace debug status
	go http.ListenAndServe("localhost:8080", nil)

	// Config
	//trustDomainConfig0 := common.NewTrustDomainConfig("t0", "localhost:9000", true, false)
	//trustDomainConfig1 := common.NewTrustDomainConfig("t1", "localhost:9001", true, false)
	trustDomainConfig0 := common.NewTrustDomainConfig("t0", "172.30.2.10:9000", true, false)
	trustDomainConfig1 := common.NewTrustDomainConfig("t1", "172.30.2.159:9000", true, false)
	trustDomainConfig2 := common.NewTrustDomainConfig("t2", "172.30.2.221:9000", true, false)
	config := common.CommonConfigFromFile("commonconfig.json")
	serverConfig := server.ServerConfigFromFile("serverconfig.json", config)
	config.TrustDomains = []*common.TrustDomainConfig{trustDomainConfig0, trustDomainConfig1, trustDomainConfig2}

	s := server.NewCentralized("s", "../pird/pir.socket", *serverConfig, nil, false)
	//s := server.NewCentralized("s", "../pird/pir.socket", *config, common.NewFollowerRpc("t1->t2", trustDomainConfig2), false)
	//s := server.NewCentralized("s", "../pird/pir.socket", *config, common.NewFollowerRpc("t0->t1", trustDomainConfig1), true)
	_ = server.NewNetworkRpc(s, 9000)

	log.Println(s)

	for {
		time.Sleep(10 * time.Second)
	}

}
