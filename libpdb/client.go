package libpdb

import (
	"github.com/ryscheng/pdb/common"
	"sync/atomic"
)

/**
 * Client interface for libpdb
 * Goroutines:
 * - 1x RequestManager.writePeriodic
 * - 1x RequestManager.readPeriodic
 */
type Client struct {
	log          *common.Logger
	name         string
	config atomic.Value //ClientConfig
	leader       common.LeaderInterface
	msgReqMan    *RequestManager
}

func NewClient(name string, config ClientConfig, leader common.LeaderInterface) *Client {
	c := &Client{}
	c.log = common.NewLogger(name)
	c.name = name
	c.config.Store(config)
	c.leader = leader

	c.msgReqMan = NewRequestManager(name, c.leader, &c.config)

	c.log.Info.Println("NewClient: starting new client - " + name)
	return c
}

/** PUBLIC METHODS (threadsafe) **/

func (c *Client) SetConfig(config ClientConfig) {
	c.config.Store(config)
}

func (c *Client) Ping() bool {
	var reply common.PingReply
	err := c.leader.Ping(&common.PingArgs{"PING"}, &reply)
	if err == nil && reply.Err == "" {
		c.log.Info.Printf("Ping success\n")
		return true
	} else {
		c.log.Warn.Printf("Ping fail: err=%v, reply=%v\n", err, reply)
		return false
	}
}

func (c *Client) CreateTopic() (*Topic, error) {
	password := ""
	handle, err := NewTopic(password)
	//@todo
	return handle, err
}

func (c *Client) Publish(data []byte) bool {
	//@todo using EnqueueWrite
	return true
}

func (c *Client) Subscribe(handle *Topic) bool {
	//@todo using EnqueueRead
	return true
}

func (c *Client) PublishTrace() uint64 {
	config := c.config.Load().(ClientConfig)
	req := &common.WriteArgs{}
	req.ReplyChan = make(chan *common.WriteReply)
	c.msgReqMan.generateRandomWrite(config, req)
	c.msgReqMan.EnqueueWrite(req)
	reply := <-req.ReplyChan
	return reply.GlobalSeqNo
}

func (c *Client) PollTrace() common.Range {
	config := c.config.Load().(ClientConfig)
	req := &common.ReadRequest{}
	req.Args = &common.ReadArgs{}
	req.ReplyChan = make(chan *common.ReadReply)
	c.msgReqMan.generateRandomRead(config, req.Args)
	c.msgReqMan.EnqueueRead(req)
	reply := <-req.ReplyChan
	return reply.GlobalSeqNo
}
