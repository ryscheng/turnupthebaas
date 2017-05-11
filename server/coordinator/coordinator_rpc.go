package coordinator

import (
	"log"
	"net/rpc"
	"os"

	"github.com/privacylab/talek/common"
)

// CoordinatorRPC is a stub for RPCs to the central coordinator server.
type CoordinatorRPC struct {
	log     *common.Logger
	name    string
	address string
	client  *rpc.Client
}

// NewCoordinatorRPC instantiates a client stub
func NewCoordinatorRPC(name string, address string) *CoordinatorRPC {
	c := &CoordinatorRPC{}
	c.log = common.NewLogger(name)
	c.name = name
	c.address = address
	c.client = nil
	return c
}

// Call implements an RPC call
func (c *CoordinatorRPC) Call(methodName string, args interface{}, reply interface{}) error {
	// Get address
	var err error

	// Setup connection
	if c.client == nil {
		c.client, err = rpc.Dial("tcp", c.address)
		if err != nil {
			c.log.Error.Printf("rpc dialing failed: %v\n", err)
			c.client = nil
			return err
		}
		//defer client.Close()
	}

	// Do RPC
	err = c.client.Call(methodName, args, reply)
	if err != nil {
		c.log.Error.Printf("rpc error: %v", err)
		return err
	}

	//l.log.Printf("%s.Call(): %v, %v, %v\n", addr, args, reply)
	return nil
}

// GetName returns the name of the leader.
func (c *CoordinatorRPC) GetInfo(_ *interface{}, reply *GetInfoReply) error {
	err := c.Call("Coordinator.GetInfo", nil, reply)
	return err
}

// GetConfig tells the client about current config.
func (c *CoordinatorRPC) GetConfig(_ *interface{}, reply *common.Config) error {
	err := c.Call("Coordinator.GetConfig", nil, reply)
	return err
}

func (c *CoordinatorRPC) Commit(args *CommitArgs, reply *CommitReply) error {
	err := c.Call("Coordinator.Commit", args, reply)
	return err
}

// GetUpdates provides the global interest vector.
func (c *CoordinatorRPC) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	err := c.Call("Coordinator.GetUpdates", args, reply)
	return err
}
