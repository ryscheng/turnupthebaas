package replica

import (
	"net/rpc"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/server/coordinator"
)

// RPCStub is a stub for RPCs to the central coordinator server.
type RPCStub struct {
	log     *common.Logger
	name    string
	address string
	client  *rpc.Client
	lastErr error
}

// NewRPCStub instantiates a client stub
func NewRPCStub(name string, address string) *RPCStub {
	c := &RPCStub{}
	c.log = common.NewLogger(name)
	c.name = name
	c.address = address
	c.client = nil // Lazily dial as necessary
	return c
}

// Close will close the RPC client
func (c *RPCStub) Close() error {
	if c.client != nil {
		c.lastErr = c.client.Close()
		c.client = nil
		return c.lastErr
	}
	return nil
}

// GetInfo returns info about this server
func (c *RPCStub) GetInfo(_ *interface{}, reply *GetInfoReply) error {
	var args interface{}
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Server.GetInfo", &args, reply)
	return c.lastErr
}

// Notify the server of a new shapshot
func (c *RPCStub) Notify(args *coordinator.NotifyArgs, reply *coordinator.NotifyReply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Server.Notify", args, reply)
	return c.lastErr
}

// Write a single message
func (c *RPCStub) Write(args *common.WriteArgs, reply *common.WriteReply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Server.Write", args, reply)
	return c.lastErr
}

// Read a batch of requests for a shard range
func (c *RPCStub) Read(args *ReadArgs, reply *ReadReply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Server.Read", args, reply)
	return c.lastErr
}
