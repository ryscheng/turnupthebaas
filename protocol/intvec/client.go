package intvec

import (
	"net/rpc"

	"github.com/privacylab/talek/common"
)

// Client is a stub for RPCs to get interest vectors
type Client struct {
	log     *common.Logger
	name    string
	address string
	client  *rpc.Client
	lastErr error
}

// NewClient instantiates a client stub
func NewClient(name string, address string) *Client {
	c := &Client{}
	c.log = common.NewLogger(name)
	c.name = name
	c.address = address
	c.client = nil // Lazily dial as necessary
	return c
}

// Close will close the RPC client
func (c *Client) Close() error {
	if c.client != nil {
		c.lastErr = c.client.Close()
		c.client = nil
		return c.lastErr
	}
	return nil
}

// GetIntVec retrieves the global interest vector
func (c *Client) GetIntVec(args *GetIntVecArgs, reply *GetIntVecReply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Server.GetIntVec", &args, reply)
	return c.lastErr
}
