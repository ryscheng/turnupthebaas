package common

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
)

// FollowerRPC is a stub for the follower RPC interface
type FollowerRPC struct {
	log          *log.Logger
	name         string
	config       *TrustDomainConfig
	methodPrefix string
	client       *rpc.Client
}

// NewFollowerRPC creates a new FollowerRPC
func NewFollowerRPC(name string, config *TrustDomainConfig) *FollowerRPC {
	f := &FollowerRPC{}
	f.log = log.New(os.Stdout, "[FollowerRPC:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
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

// Call calls an RPC method.
func (f *FollowerRPC) Call(methodName string, args interface{}, reply interface{}) error {
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
		f.log.Printf("rpc error: %v", err)
		return err
	}

	//f.log.Printf("%s.Call(): %v, %v, %v\n", addr, args, reply)
	return nil
}

// GetName returns the name of the follower.
func (f *FollowerRPC) GetName(_ *interface{}, reply *string) error {
	*reply = f.name
	return nil
}

func (f *FollowerRPC) Write(args *WriteArgs, reply *WriteReply) error {
	//f.log.Printf("Write: enter\n")
	err := f.Call(f.methodPrefix+".Write", args, reply)
	return err
}

// NextEpoch signals a new write epoch
func (f *FollowerRPC) NextEpoch(args *uint64, reply *interface{}) error {
	err := f.Call(f.methodPrefix+".NextEpoch", args, reply)
	return err
}

// BatchRead performs a set of PIR reads.
func (f *FollowerRPC) BatchRead(args *BatchReadRequest, reply *BatchReadReply) error {
	//f.log.Printf("BatchRead: enter\n")
	err := f.Call(f.methodPrefix+".BatchRead", args, reply)
	return err
}

// GetUpdates provies the most recent set of global interest vector changes.
func (f *FollowerRPC) GetUpdates(args *GetUpdatesArgs, reply *GetUpdatesReply) error {
	//f.log.Printf("GetUpdates: enter\n")
	err := f.Call(f.methodPrefix+".GetUpdates", args, reply)
	return err
}
