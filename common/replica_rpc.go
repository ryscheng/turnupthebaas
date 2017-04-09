package common

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
)

// ReplicaRPC is a stub for the replica RPC interface
type ReplicaRPC struct {
	log          *log.Logger
	name         string
	config       *TrustDomainConfig
	methodPrefix string
	client       *rpc.Client
}

// NewReplicaRPC creates a new ReplicaRPC
func NewReplicaRPC(name string, config *TrustDomainConfig) *ReplicaRPC {
	r := &ReplicaRPC{}
	r.log = log.New(os.Stdout, "[ReplicaRPC:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	r.name = name
	r.config = config
	r.client = nil
	if r.config.IsDistributed {
		r.methodPrefix = "Frontend"
	} else {
		r.methodPrefix = "Centralized"
	}

	return r
}

// Call calls an RPC method.
func (r *ReplicaRPC) Call(methodName string, args interface{}, reply interface{}) error {
	// Get address
	var err error
	addr, okAddr := r.config.GetAddress()
	if !okAddr {
		return fmt.Errorf("No address available")
	}

	// Setup connection
	if r.client == nil {
		r.client, err = rpc.Dial("tcp", addr)
		if err != nil {
			r.log.Printf("rpc dialing failed: %v\n", err)
			r.client = nil
			return err
		}
		//defer client.Close()
	}

	// Do RPC
	err = r.client.Call(methodName, args, reply)
	if err != nil {
		r.log.Printf("rpc error: %v", err)
		return err
	}

	//f.log.Printf("%s.Call(): %v, %v, %v\n", addr, args, reply)
	return nil
}

func (r *ReplicaRPC) Write(args *ReplicaWriteArgs, reply *ReplicaWriteReply) error {
	//f.log.Printf("Write: enter\n")
	err := r.Call(r.methodPrefix+".Write", args, reply)
	return err
}

// BatchRead performs a set of PIR reads.
func (r *ReplicaRPC) BatchRead(args *BatchReadRequest, reply *BatchReadReply) error {
	//f.log.Printf("BatchRead: enter\n")
	err := r.Call(r.methodPrefix+".BatchRead", args, reply)
	return err
}

// GetUpdates provies the most recent set of global interest vector changes.
func (r *ReplicaRPC) GetUpdates(args *GetUpdatesArgs, reply *GetUpdatesReply) error {
	//f.log.Printf("GetUpdates: enter\n")
	err := r.Call(r.methodPrefix+".GetUpdates", args, reply)
	return err
}
