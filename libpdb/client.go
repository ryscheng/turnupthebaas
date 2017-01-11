package libpdb

import (
	"github.com/ryscheng/pdb/common"
	"sync"
	"sync/atomic"
)

/**
 * Client interface for libpdb
 * Goroutines:
 * - 1x RequestManager.writePeriodic
 * - 1x RequestManager.readPeriodic
 */
type Client struct {
	log       *common.Logger
	name      string
	config    atomic.Value //ClientConfig
	leader    common.LeaderInterface
	msgReqMan *RequestManager

	subscriptions     []Subscription
	pendingRequest    *common.ReadRequest
	pendingRequestSub *RequestResponder
	subscriptionMutex sync.Mutex
}

//TODO: client needs to know the different trust domains security parameters.
func NewClient(name string, config ClientConfig, leader common.LeaderInterface) *Client {
	c := &Client{}
	c.log = common.NewLogger(name)
	c.name = name
	c.config.Store(config)
	c.leader = leader

	c.msgReqMan = NewRequestManager(name, c.leader, &c.config)
	c.msgReqMan.SetReadGenerator(c)
	c.subscriptionMutex = sync.Mutex{}

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

func (c *Client) Publish(handle *Topic, data []byte) error {
	config := c.config.Load().(ClientConfig)
	write_args, err := handle.GeneratePublish(config, seqNo, data)
	if error != nil {
		return err
	}

	c.msgReqMan.EnqueueWrite(write_args)
	return true
}

func (c *Client) Subscribe(handle *Subscription) bool {
	// Check if already subscribed.
	c.subscriptionMutex.Lock()
	for x := range c.subscriptions {
		if x == handle {
			c.subscriptionMutex.Unlock()
			return false
		}
	}
	c.subscriptions = append(c.subscriptions, handle)
	c.subscriptionMutex.Unlock()

	return true
}

func (c *Client) Unsubscribe(handle *Subscription) bool {
	c.subscriptionMutex.Lock()
	for i := 0; i < len(c.subscriptions); i++ {
		if c.subscriptions[i] == handle {
			c.subscriptions[i] = c.subscriptionMutex[len(c.subscriptions)-1]
			c.subscriptions = c.subscriptions[:len(c.subscriptions)-1]
			c.subscriptionMutex.Unlock()
			return true
		}
	}
	c.subscriptionMutex.Unlock()
	return false
}

// Implement RequestGenerator interface for the request manager
func (c *Client) NextRequest() *common.ReadRequest {
	c.subscriptionMutex.Lock()
	if c.pendingRequest != nil {
		rec := c.pendingRequest
		rr := c.pendingRequestSub
		c.pendingRequest = nil
		c.subscriptionMutex.Unlock()
		return req, rr
	}

	if len(c.subscriptions) > 0 {
		nextTopic := c.subscriptions[0]
		c.subscriptions = c.subscriptions[1:]
		c.subscriptions = append(c.subscriptions, nextTopic)

		ra1, ra2, err := nextTopic.generatePoll(config, seqNo)
		if err {
			c.subscriptionMutex.Unlock()
			c.log.Error(err)
			return nil
		}
		c.pendingRequest = ra2
		c.pendingRequestSub = nextTopic
		c.subscriptionMutex.Unlock()
		return ra1, nextTopic
	}
	c.subscriptionMutex.Unlock()
	return nil
}

// Debug only. For learning latencies.
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
