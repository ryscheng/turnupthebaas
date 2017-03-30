package libtalek

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/drbg"
)

// Client represents a connection to the Talek system. Typically created with
// NewClient, the object manages requests, both reads an writes.
type Client struct {
	log    *common.Logger
	name   string
	config atomic.Value //ClientConfig
	dead   int32
	leader common.LeaderInterface

	handles       []Handle
	pendingWrites chan *common.WriteArgs
	pendingReads  chan request
	handleMutex   sync.Mutex

	lastSeqNo uint64
}

type request struct {
	*common.ReadArgs
	*Handle
}

// NewClient creates a Talek client for reading and writing metadata-protected messages.
func NewClient(name string, config ClientConfig, leader common.LeaderInterface) *Client {
	c := &Client{}
	c.log = common.NewLogger(name)
	c.name = name
	c.config.Store(config)
	c.leader = leader

	//todo: should channel capacity be smarter?
	c.pendingReads = make(chan request, 5)
	c.pendingWrites = make(chan *common.WriteArgs, 5)

	go c.readPeriodic()
	go c.writePeriodic()

	c.log.Info.Println("NewClient: starting new client - " + name)
	return c
}

/** PUBLIC METHODS (threadsafe) **/

// SetConfig allows updating the configuraton of a Client, e.g. if server memebership
// or speed characteristics for the system are changed.
func (c *Client) SetConfig(config ClientConfig) {
	c.config.Store(config)
}

// Kill stops client processing. This allows for graceful shutdown or suspension of requests.
func (c *Client) Kill() {
	atomic.StoreInt32(&c.dead, 1)
}

// Ping will perform an on-thread ping of the Talek system, allowing the client
// to validate that it is connected to the Talek system, and check the latency
// of connection to the server.
func (c *Client) Ping() bool {
	var reply common.PingReply
	err := c.leader.Ping(&common.PingArgs{Msg: "PING"}, &reply)
	if err == nil && reply.Err == "" {
		c.log.Info.Printf("Ping success\n")
		return true
	}
	c.log.Warn.Printf("Ping fail: err=%v, reply=%v\n", err, reply)
	return false
}

// MaxLength returns the maximum allowed message the client can Publish.
// TODO: support messages spanning multiple data items.
func (c *Client) MaxLength() int {
	config := c.config.Load().(ClientConfig)
	return config.DataSize
}

// Publish a new message to the end of a topic.
func (c *Client) Publish(handle *Topic, data []byte) error {
	config := c.config.Load().(ClientConfig)

	if len(data) > config.DataSize-PublishingOverhead {
		return errors.New("message too long")
	} else if len(data) < config.DataSize-PublishingOverhead {
		allocation := make([]byte, config.DataSize-PublishingOverhead)
		copy(allocation[:], data)
		data = allocation
	}

	writeArgs, err := handle.GeneratePublish(config.Config, data)
	c.log.Info.Printf("Wrote %v(%d) to %d,%d.", writeArgs.Data[0:4], len(writeArgs.Data), writeArgs.Bucket1, writeArgs.Bucket2)
	if err != nil {
		return err
	}

	c.pendingWrites <- writeArgs
	return nil
}

// Poll handles to updates on a given log.
// When done reading message,s the the channel can be closed via the Done
// method.
func (c *Client) Poll(handle *Handle) chan []byte {
	// Check if already polling.
	c.handleMutex.Lock()
	for x := range c.handles {
		if &c.handles[x] == handle {
			c.handleMutex.Unlock()
			return nil
		}
	}
	if handle.updates == nil {
		initHandle(handle)
	}
	c.handles = append(c.handles, *handle)
	c.handleMutex.Unlock()

	return handle.updates
}

// Done unsubscribes a Handle from being Polled for new items.
func (c *Client) Done(handle *Handle) bool {
	c.handleMutex.Lock()
	for i := 0; i < len(c.handles); i++ {
		if &c.handles[i] == handle {
			c.handles[i] = c.handles[len(c.handles)-1]
			c.handles = c.handles[:len(c.handles)-1]
			c.handleMutex.Unlock()
			return true
		}
	}
	c.handleMutex.Unlock()
	return false
}

/** Private methods **/
func (c *Client) writePeriodic() {
	var req *common.WriteArgs

	for atomic.LoadInt32(&c.dead) == 0 {
		reply := common.WriteReply{}
		conf := c.config.Load().(ClientConfig)
		select {
		case req = <-c.pendingWrites:
			break
		default:
			req = c.generateRandomWrite(conf)
		}
		err := c.leader.Write(req, &reply)
		if err != nil {
			reply.Err = err.Error()
		}
		if reply.GlobalSeqNo > c.lastSeqNo {
			c.lastSeqNo = reply.GlobalSeqNo
		}
		if req.ReplyChan != nil {
			req.ReplyChan <- &reply
		}
		//TODO: switch to poisson
		time.Sleep(conf.WriteInterval)
	}
}

func (c *Client) readPeriodic() {
	var req request

	for atomic.LoadInt32(&c.dead) == 0 {
		reply := common.ReadReply{}
		conf := c.config.Load().(ClientConfig)
		select {
		case req = <-c.pendingReads:
			break
		default:
			req = c.nextRequest(&conf)
		}
		encreq, err := req.ReadArgs.Encode(conf.TrustDomains)
		if err != nil {
			reply.Err = err.Error()
		} else {
			err := c.leader.Read(&encreq, &reply)
			if err != nil {
				reply.Err = err.Error()
			}
		}
		if reply.GlobalSeqNo.End > c.lastSeqNo {
			c.lastSeqNo = reply.GlobalSeqNo.End
		}
		if req.Handle != nil {
			req.Handle.OnResponse(req.ReadArgs, &reply, uint(conf.DataSize))
		}
		time.Sleep(conf.ReadInterval)
	}
}

func (c *Client) generateRandomWrite(config ClientConfig) *common.WriteArgs {
	args := &common.WriteArgs{}
	var max big.Int
	b1, _ := rand.Int(rand.Reader, max.SetUint64(config.NumBuckets))
	b2, _ := rand.Int(rand.Reader, max.SetUint64(config.NumBuckets))
	args.Bucket1 = b1.Uint64()
	args.Bucket2 = b2.Uint64()
	args.Data = make([]byte, config.Config.DataSize, config.Config.DataSize)
	rand.Read(args.Data)
	return args
}

func (c *Client) generateRandomRead(config *ClientConfig) *common.ReadArgs {
	args := &common.ReadArgs{}
	vectorSize := uint32((config.Config.NumBuckets+7)/8 + 1)
	args.TD = make([]common.PirArgs, len(config.TrustDomains), len(config.TrustDomains))
	for i := 0; i < len(args.TD); i++ {
		args.TD[i].RequestVector = make([]byte, vectorSize, vectorSize)
		rand.Read(args.TD[i].RequestVector)
		seed, err := drbg.NewSeed()
		if err != nil {
			c.log.Error.Fatalf("Error creating random seed: %v\n", err)
		}
		args.TD[i].PadSeed, err = seed.MarshalBinary()
		if err != nil {
			c.log.Error.Fatalf("Failed to marshal seed: %v\n", err)
		}
	}
	return args
}

func (c *Client) nextRequest(config *ClientConfig) request {
	c.handleMutex.Lock()

	if len(c.handles) > 0 {
		nextTopic := c.handles[0]
		c.handles = c.handles[1:]
		c.handles = append(c.handles, nextTopic)

		ra1, ra2, err := nextTopic.generatePoll(config, c.lastSeqNo)
		if err != nil {
			c.handleMutex.Unlock()
			c.log.Error.Fatal(err)
			return request{c.generateRandomRead(config), nil}
		}
		c.pendingReads <- request{ra2, &nextTopic}
		c.handleMutex.Unlock()
		return request{ra1, &nextTopic}
	}
	c.handleMutex.Unlock()

	return request{c.generateRandomRead(config), nil}
}
