package coordinator

import "github.com/privacylab/talek/common"

// Client is a stub for RPCs to the central coordinator server.
type Client struct {
	log     *common.Logger
	name    string
	address string
	lastErr error
}

// NewClient instantiates a client stub
func NewClient(name string, address string) *Client {
	c := &Client{}
	c.log = common.NewLogger(name)
	c.name = name
	c.address = address
	return c
}

// Close will close the RPC client
func (c *Client) Close() error {
	return nil
}

// GetInfo returns info about this server
func (c *Client) GetInfo(_ *interface{}, reply *GetInfoReply) error {
	var args interface{}
	c.lastErr = common.RPCCall(nil, c.address, "Coordinator.GetInfo", &args, reply)
	return c.lastErr
}

// GetCommonConfig returns the current config.
func (c *Client) GetCommonConfig(_ *interface{}, reply *common.Config) error {
	var args interface{}
	c.lastErr = common.RPCCall(nil, c.address, "Coordinator.GetCommonConfig", &args, reply)
	return c.lastErr
}

// GetLayout provides the layout for a shard
func (c *Client) GetLayout(args *GetLayoutArgs, reply *GetLayoutReply) error {
	c.lastErr = common.RPCCall(nil, c.address, "Coordinator.GetLayout", args, reply)
	return c.lastErr
}

// GetIntVec provides the global interest vector
func (c *Client) GetIntVec(args *GetIntVecArgs, reply *GetIntVecReply) error {
	c.lastErr = common.RPCCall(nil, c.address, "Coordinator.GetIntVec", args, reply)
	return c.lastErr
}

// Commit a set of Writes
func (c *Client) Commit(args *CommitArgs, reply *CommitReply) error {
	c.lastErr = common.RPCCall(nil, c.address, "Coordinator.Commit", args, reply)
	return c.lastErr
}
