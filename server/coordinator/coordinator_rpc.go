package coordinator

import (
	"net/rpc"

	"github.com/privacylab/talek/common"
)

// CoordinatorRPC is a stub for RPCs to the central coordinator server.
type CoordinatorRPC struct {
	log     *common.Logger
	name    string
	address string
	client  *rpc.Client
	lastErr error
}

// NewCoordinatorRPC instantiates a client stub
func NewCoordinatorRPC(name string, address string) *CoordinatorRPC {
	c := &CoordinatorRPC{}
	c.log = common.NewLogger(name)
	c.name = name
	c.address = address
	c.client = nil // Lazily dial as necessary
	return c
}

// Close will close the RPC client
func (c *CoordinatorRPC) Close() error {
	if c.client != nil {
		c.lastErr = c.client.Close()
		c.client = nil
		return c.lastErr
	}
	return nil
}

// GetName returns the name of the leader.
func (c *CoordinatorRPC) GetInfo(_ *interface{}, reply *GetInfoReply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Coordinator.GetInfo", nil, reply)
	return c.lastErr
}

// GetConfig tells the client about current config.
func (c *CoordinatorRPC) GetConfig(_ *interface{}, reply *common.Config) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Coordinator.GetConfig", nil, reply)
	return c.lastErr
}

func (c *CoordinatorRPC) Commit(args *CommitArgs, reply *CommitReply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Coordinator.Commit", args, reply)
	return c.lastErr
}

// GetUpdates provides the global interest vector.
func (c *CoordinatorRPC) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Coordinator.GetUpdates", args, reply)
	return c.lastErr
}
