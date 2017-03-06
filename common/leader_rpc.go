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
	client       *rpc.Client
}

func NewLeaderRpc(name string, config *TrustDomainConfig) *LeaderRpc {
	l := &LeaderRpc{}
	l.log = log.New(os.Stdout, "[LeaderRpc:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	l.name = name
	l.config = config
	l.client = nil
	if l.config.IsDistributed {
		l.methodPrefix = "Frontend"
	} else {
		l.methodPrefix = "Centralized"
	}

	return l
}

func (l *LeaderRpc) Call(methodName string, args interface{}, reply interface{}) error {
	// Get address
	var err error
	addr, okAddr := l.config.GetAddress()
	if !okAddr {
		return fmt.Errorf("No address available")
	}

	// Setup connection
	if l.client == nil {
		l.client, err = rpc.Dial("tcp", addr)
		if err != nil {
			l.log.Printf("rpc dialing failed: %v\n", err)
			l.client = nil
			return err
		}
		//defer client.Close()
	}

	// Do RPC
	err = l.client.Call(methodName, args, reply)
	if err != nil {
		l.log.Printf("rpc error: %v", err)
		return err
	}

	//l.log.Printf("%s.Call(): %v, %v, %v\n", addr, args, reply)
	return nil
}

func (l *LeaderRpc) GetName(_ *interface{}, reply *string) error {
	*reply = l.name
	return nil
}

func (l *LeaderRpc) Ping(args *PingArgs, reply *PingReply) error {
	//l.log.Printf("Ping: enter\n")
	err := l.Call(l.methodPrefix+".Ping", args, reply)
	return err
}

func (l *LeaderRpc) Write(args *WriteArgs, reply *WriteReply) error {
	//l.log.Printf("Write: enter\n")
	err := l.Call(l.methodPrefix+".Write", args, reply)
	return err
}

func (l *LeaderRpc) Read(args *EncodedReadArgs, reply *ReadReply) error {
	//l.log.Printf("Read: enter\n")
	err := l.Call(l.methodPrefix+".Read", args, reply)
	return err
}

func (l *LeaderRpc) GetUpdates(args *GetUpdatesArgs, reply *GetUpdatesReply) error {
	//l.log.Printf("GetUpdates: enter\n")
	err := l.Call(l.methodPrefix+".GetUpdates", args, reply)
	return err
}
