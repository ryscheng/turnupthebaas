package common

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
)

// FrontendRPC is a stub for RPCs to the talek server.
type FrontendRPC struct {
	log          *log.Logger
	name         string
	config       *TrustDomainConfig
	methodPrefix string
	client       *rpc.Client
}

// NewFrontendRPC instantiates a LeaderRPC stub
func NewFrontendRPC(name string, config *TrustDomainConfig) *FrontendRPC {
	f := &FrontendRPC{}
	f.log = log.New(os.Stdout, "[FrontendRPC:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	f.name = name
	f.config = config
	f.client = nil
	f.methodPrefix = "Frontend"

	return f
}

// Call implements an RPC call
func (f *FrontendRPC) Call(methodName string, args interface{}, reply interface{}) error {
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

	//l.log.Printf("%s.Call(): %v, %v, %v\n", addr, args, reply)
	return nil
}

// GetName returns the name of the leader.
func (f *FrontendRPC) GetName(_ *interface{}, reply *string) error {
	*reply = f.name
	return nil
}

// GetConfig tells the client about current config.
func (f *FrontendRPC) GetConfig(_ *interface{}, reply *Config) error {
	err := f.Call(f.methodPrefix+".Config", nil, reply)
	return err
}

func (f *FrontendRPC) Write(args *WriteArgs, reply *WriteReply) error {
	//l.log.Printf("Write: enter\n")
	err := f.Call(f.methodPrefix+".Write", args, reply)
	return err
}

func (f *FrontendRPC) Read(args *EncodedReadArgs, reply *ReadReply) error {
	//l.log.Printf("Read: enter\n")
	err := f.Call(f.methodPrefix+".Read", args, reply)
	return err
}

// GetUpdates provides the global interest vector.
func (f *FrontendRPC) GetUpdates(args *GetUpdatesArgs, reply *GetUpdatesReply) error {
	//l.log.Printf("GetUpdates: enter\n")
	err := f.Call(f.methodPrefix+".GetUpdates", args, reply)
	return err
}
