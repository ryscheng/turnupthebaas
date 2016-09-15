package common

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
)

type TrustDomainRef struct {
	log          *log.Logger
	config       *TrustDomainConfig
	methodPrefix string
}

func NewTrustDomainRef(name string, config *TrustDomainConfig) *TrustDomainRef {
	t := &TrustDomainRef{}
	t.log = log.New(os.Stdout, "[TrustDomainRef:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	t.config = config
	if t.config.IsDistributed() {
		t.methodPrefix = "Frontend"
	} else {
		t.methodPrefix = "Centralized"
	}

	return t
}

func (t *TrustDomainRef) Call(methodName string, args interface{}, reply interface{}) error {
	// Get address
	addr, okAddr := t.config.GetAddress()
	if !okAddr {
		return fmt.Errorf("No address available")
	}

	// Setup connection
	client, errDial := rpc.Dial("tcp", addr)
	if errDial != nil {
		log.Printf("rpc dialing failed: %v\n", errDial)
		return errDial
	}
	defer client.Close()

	// Do RPC
	errCall := client.Call(methodName, args, reply)
	if errCall != nil {
		log.Printf("rpc error:", errCall)
		return errCall
	}

	t.log.Printf("%s.Write(): %v, %v, %v\n", addr, args, reply)
	return nil
}

func (t *TrustDomainRef) Ping() (*PingReply, error) {
	t.log.Printf("Ping: enter\n")
	args := &PingArgs{"PING"}
	var reply PingReply
	err := t.Call(t.methodPrefix+".Ping", args, &reply)
	return &reply, err
}

func (t *TrustDomainRef) Write(args *WriteArgs) (*WriteReply, error) {
	t.log.Printf("Write: enter\n")
	var reply WriteReply
	err := t.Call(t.methodPrefix+".Write", args, &reply)
	return &reply, err
}
