package common

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
)

type FollowerRpc struct {
	log          *log.Logger
	name         string
	config       *TrustDomainConfig
	methodPrefix string
	client       *rpc.Client
}

func NewFollowerRpc(name string, config *TrustDomainConfig) *FollowerRpc {
	f := &FollowerRpc{}
	f.log = log.New(os.Stdout, "[FollowerRpc:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	f.name = name
	f.config = config
	f.client = nil
	if f.config.IsDistributed {
		f.methodPrefix = "Frontend"
	} else {
		f.methodPrefix = "Centralized"
	}

	return f
}

func (f *FollowerRpc) Call(methodName string, args interface{}, reply interface{}) error {
	// Get address
	var err error
	addr, okAddr := f.config.GetAddress()
	if !okAddr {
		return fmt.Errorf("No address available")
	}

	// Setup connection
	if f.client == nil {
		f.client, err = rpc.Dial("tcp", addr)
		if err != nil {
			f.log.Printf("rpc dialing failed: %v\n", err)
			f.client = nil
			return err
		}
		//defer client.Close()
	}

	// Do RPC
	err = f.client.Call(methodName, args, reply)
	if err != nil {
		f.log.Printf("rpc error:", err)
		return err
	}

	//f.log.Printf("%s.Call(): %v, %v, %v\n", addr, args, reply)
	return nil
}

func (f *FollowerRpc) GetName() string {
	return f.name
}

func (f *FollowerRpc) Ping(args *PingArgs, reply *PingReply) error {
	//f.log.Printf("Ping: enter\n")
	err := f.Call(f.methodPrefix+".Ping", args, reply)
	return err
}

func (f *FollowerRpc) Write(args *WriteArgs, reply *WriteReply) error {
	//f.log.Printf("Write: enter\n")
	err := f.Call(f.methodPrefix+".Write", args, reply)
	return err
}

func (f *FollowerRpc) BatchRead(args *BatchReadRequest, reply *BatchReadReply) error {
	//f.log.Printf("BatchRead: enter\n")
	err := f.Call(f.methodPrefix+".BatchRead", args, reply)
	return err
}

func (f *FollowerRpc) GetUpdates(args *GetUpdatesArgs, reply *GetUpdatesReply) error {
	//f.log.Printf("GetUpdates: enter\n")
	err := f.Call(f.methodPrefix+".GetUpdates", args, reply)
	return err
}
