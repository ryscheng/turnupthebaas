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
	globalConfig atomic.Value //common.GlobalConfig
	leader       common.LeaderInterface
	msgReqMan    *RequestManager
}

func NewClient(name string, globalConfig common.GlobalConfig, leader common.LeaderInterface) *Client {
	c := &Client{}
	c.log = common.NewLogger(name)
	c.name = name
	c.globalConfig.Store(globalConfig)
	c.leader = leader

	c.msgReqMan = NewRequestManager(name, c.leader, &c.globalConfig)

	c.log.Info.Println("NewClient: starting new client - " + name)
	return c
}

/** PUBLIC METHODS (threadsafe) **/

func (c *Client) SetGlobalConfig(globalConfig common.GlobalConfig) {
	c.globalConfig.Store(globalConfig)
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
