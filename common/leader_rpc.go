package common

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
)

type LeaderRpc struct {
	log          *log.Logger
	name         string
	config       *TrustDomainConfig
	methodPrefix string
}

func NewLeaderRpc(name string, config *TrustDomainConfig) *LeaderRpc {
	l := &LeaderRpc{}
	l.log = log.New(os.Stdout, "[LeaderRpc:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	l.name = name
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

	//l.log.Printf("%s.Call(): %v, %v, %v\n", addr, args, reply)
	return nil
}

func (l *LeaderRpc) GetName() string {
	return l.name
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

func (l *LeaderRpc) Read(args *ReadArgs, reply *ReadReply) error {
	l.log.Printf("Read: enter\n")
	err := l.Call(l.methodPrefix+".Read", args, reply)
	return err
}

func (l *LeaderRpc) GetUpdates(args *GetUpdatesArgs, reply *GetUpdatesReply) error {
	l.log.Printf("GetUpdates: enter\n")
	err := l.Call(l.methodPrefix+".GetUpdates", args, reply)
	return err
}
