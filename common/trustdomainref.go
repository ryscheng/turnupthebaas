package common

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
)

type TrustDomainRef struct {
	log    *log.Logger
	config *TrustDomainConfig
}

func NewTrustDomainRef(name string, config *TrustDomainConfig) *TrustDomainRef {
	t := &TrustDomainRef{}
	t.log = log.New(os.Stdout, "[TrustDomainRef:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	t.config = config

	return t
}

func Call(addr string, methodName string, args interface{}, reply interface{}) error {
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		log.Printf("rpc dialing failed: %v\n", err)
		return err
	}
	defer client.Close()

	err = client.Call(methodName, args, reply)
	if err != nil {
		log.Printf("rpc error:", err)
		return err
	}
	return nil
}

func (t *TrustDomainRef) Ping() (error, *PingReply) {
	t.log.Printf("Ping: enter\n")
	args := &PingArgs{"PING"}
	var reply PingReply
	addr, ok := t.config.GetAddress()
	if !ok {
		return fmt.Errorf("No address available"), nil
	}
	err := Call(addr, "Frontend.Ping", args, &reply)

	t.log.Printf("%s.Ping(): %v, %v, %v\n", addr, args, reply, err)
	return err, &reply
}
