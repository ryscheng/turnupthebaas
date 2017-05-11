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

// NewCoordinatorRPC instantiates a client stub
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

// GetName returns the name of the leader.
func (c *RPCStub) GetInfo(_ *interface{}, reply *GetInfoReply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Coordinator.GetInfo", nil, reply)
	return c.lastErr
}

// GetConfig tells the client about current config.
func (c *RPCStub) GetConfig(_ *interface{}, reply *common.Config) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Coordinator.GetConfig", nil, reply)
	return c.lastErr
}

func (c *RPCStub) Commit(args *CommitArgs, reply *CommitReply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Coordinator.Commit", args, reply)
	return c.lastErr
}

// GetUpdates provides the global interest vector.
func (c *RPCStub) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Coordinator.GetUpdates", args, reply)
	return c.lastErr
}
