package common

import (
	"log"
	"net/http"
	"os"
)

// FrontendRPC is a stub for RPCs to the talek server.
type FrontendRPC struct {
	log          *log.Logger
	name         string
	address      string
	methodPrefix string
	*http.Client
}

// NewFrontendRPC instantiates a LeaderRPC stub
func NewFrontendRPC(name string, address string) *FrontendRPC {
	f := &FrontendRPC{}
	f.log = log.New(os.Stdout, "[FrontendRPC:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	f.name = name
	f.address = address
	f.methodPrefix = "Frontend"

	return f
}

// GetName returns the name of the leader.
func (f *FrontendRPC) GetName(_ *interface{}, reply *string) error {
	*reply = f.name
	return nil
}

// GetConfig tells the client about current config.
func (f *FrontendRPC) GetConfig(_ *interface{}, reply *Config) error {
	var args interface{}
	err := RPCCall(f.Client, f.address, f.methodPrefix+".GetConfig", &args, reply)
	return err
}

func (f *FrontendRPC) Write(args *WriteArgs, reply *WriteReply) error {
	//l.log.Printf("Write: enter\n")
	err := RPCCall(f.Client, f.address, f.methodPrefix+".Write", args, reply)
	return err
}

func (f *FrontendRPC) Read(args *EncodedReadArgs, reply *ReadReply) error {
	//l.log.Printf("Read: enter\n")
	err := RPCCall(f.Client, f.address, f.methodPrefix+".Read", args, reply)
	return err
}

// GetUpdates provides the global interest vector.
func (f *FrontendRPC) GetUpdates(args *GetUpdatesArgs, reply *GetUpdatesReply) error {
	//l.log.Printf("GetUpdates: enter\n")
	err := RPCCall(f.Client, f.address, f.methodPrefix+".GetUpdates", args, reply)
	return err
}
