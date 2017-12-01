package common

import (
	"log"
	"os"
)

// ReplicaRPC is a stub for the replica RPC interface
type ReplicaRPC struct {
	log          *log.Logger
	name         string
	address      string
	methodPrefix string
}

// NewReplicaRPC creates a new ReplicaRPC
func NewReplicaRPC(name string, config *TrustDomainConfig) *ReplicaRPC {
	r := &ReplicaRPC{}
	r.log = log.New(os.Stdout, "[ReplicaRPC:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	r.name = name
	addr, ok := config.GetAddress()
	if ok {
		r.address = addr
	} else {
		return nil
	}
	r.methodPrefix = "Replica"

	return r
}

func (r *ReplicaRPC) Write(args *ReplicaWriteArgs, reply *ReplicaWriteReply) error {
	//f.log.Printf("Write: enter\n")
	err := RPCCall(r.address, r.methodPrefix+".Write", args, reply)
	return err
}

// BatchRead performs a set of PIR reads.
func (r *ReplicaRPC) BatchRead(args *BatchReadRequest, reply *BatchReadReply) error {
	//f.log.Printf("BatchRead: enter\n")
	err := RPCCall(r.address, r.methodPrefix+".BatchRead", args, reply)
	return err
}

// GetUpdates provies the most recent set of global interest vector changes.
func (r *ReplicaRPC) GetUpdates(args *GetUpdatesArgs, reply *GetUpdatesReply) error {
	//f.log.Printf("GetUpdates: enter\n")
	err := RPCCall(r.address, r.methodPrefix+".GetUpdates", args, reply)
	return err
}
