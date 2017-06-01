package fedomain

import (
	"net/rpc"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/protocol/layout"
	"github.com/privacylab/talek/protocol/notify"
)

// Client is a stub for RPCs
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
	c.lastErr = nil
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

// GetInfo returns info about this server
func (c *Client) GetInfo(_ *interface{}, reply *GetInfoReply) error {
	var args interface{}
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Server.GetInfo", &args, reply)
	return c.lastErr
}

// GetLayout retrieves a layout (potentially partial)
func (c *Client) GetLayout(args *layout.GetLayoutArgs, reply *layout.GetLayoutReply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Server.GetLayout", args, reply)
	return c.lastErr
}

// Notify the server of a new shapshot
func (c *Client) Notify(args *notify.Args, reply *notify.Reply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Server.Notify", args, reply)
	return c.lastErr
}

// Write a single message
func (c *Client) Write(args *common.WriteArgs, reply *common.WriteReply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Server.Write", args, reply)
	return c.lastErr
}

// EncPIR for an encrypted batch PIR request for a shard range
func (c *Client) EncPIR(args *EncPIRArgs, reply *EncPIRReply) error {
	c.client, c.lastErr = common.RPCCall(c.client, c.address, "Server.EncPIR", args, reply)
	return c.lastErr
}
