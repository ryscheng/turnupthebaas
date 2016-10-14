package common

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
)

type FollowerRpc struct {
	log          *log.Logger
	config       *TrustDomainConfig
	methodPrefix string
}

func NewFollowerRpc(name string, config *TrustDomainConfig) *FollowerRpc {
	f := &FollowerRpc{}
	f.log = log.New(os.Stdout, "[FollowerRpc:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	f.config = config
	if f.config.IsDistributed() {
		f.methodPrefix = "Frontend"
	} else {
		f.methodPrefix = "Centralized"
	}

	return f
}

func (f *FollowerRpc) Call(methodName string, args interface{}, reply interface{}) error {
	// Get address
	addr, okAddr := f.config.GetAddress()
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

	f.log.Printf("%s.Write(): %v, %v, %v\n", addr, args, reply)
	return nil
}

func (f *FollowerRpc) Ping(args *PingArgs, reply *PingReply) error {
	f.log.Printf("Ping: enter\n")
	err := f.Call(f.methodPrefix+".Ping", args, reply)
	return err
}

func (f *FollowerRpc) Write(args *WriteArgs, reply *WriteReply) error {
	f.log.Printf("Write: enter\n")
	err := f.Call(f.methodPrefix+".Write", args, reply)
	return err
}
