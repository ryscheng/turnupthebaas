package coordinator

import (
	"net/rpc"

	"github.com/privacylab/talek/common"
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

// GetCommonConfig returns the current config.
func (c *RPCStub) GetCommonConfig(_ *interface{}, reply *common.Config) error {
	var args interface{}
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Server.GetCommonConfig", &args, reply)
	return c.lastErr
}

// GetLayout provides the layout for a shard
func (c *RPCStub) GetLayout(args *GetLayoutArgs, reply *GetLayoutReply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Server.GetLayout", args, reply)
	return c.lastErr
}

// GetIntVec provides the global interest vector
func (c *RPCStub) GetIntVec(args *GetIntVecArgs, reply *GetIntVecReply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Server.GetIntVec", args, reply)
	return c.lastErr
}

// Commit a set of Writes
func (c *RPCStub) Commit(args *CommitArgs, reply *CommitReply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Server.Commit", args, reply)
	return c.lastErr
}
