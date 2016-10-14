package common

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
)

type LeaderRpc struct {
	log          *log.Logger
	config       *TrustDomainConfig
	methodPrefix string
}

func NewLeaderRpc(name string, config *TrustDomainConfig) *LeaderRpc {
	l := &LeaderRpc{}
	l.log = log.New(os.Stdout, "[LeaderRpc:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	l.config = config
	if l.config.IsDistributed() {
		l.methodPrefix = "Frontend"
	} else {
		l.methodPrefix = "Centralized"
	}

	return l
}

func (l *LeaderRpc) Call(methodName string, args interface{}, reply interface{}) error {
	// Get address
	addr, okAddr := l.config.GetAddress()
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

	l.log.Printf("%s.Write(): %v, %v, %v\n", addr, args, reply)
	return nil
}

func (l *LeaderRpc) Ping(args *PingArgs, reply *PingReply) error {
	l.log.Printf("Ping: enter\n")
	err := l.Call(l.methodPrefix+".Ping", args, reply)
	return err
}

func (l *LeaderRpc) Write(args *WriteArgs, reply *WriteReply) error {
	l.log.Printf("Write: enter\n")
	err := l.Call(l.methodPrefix+".Write", args, reply)
	return err
}
