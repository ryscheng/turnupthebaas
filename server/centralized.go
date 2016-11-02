package server

import (
	"github.com/ryscheng/pdb/common"
	"log"
	"os"
	"sync/atomic"
)

const BATCH_SIZE = 1

type Centralized struct {
	/** Private State **/
	// Static
	log      *log.Logger
	name     string
	follower common.FollowerInterface
	isLeader bool

	// Thread-safe
	globalConfig atomic.Value //common.GlobalConfig
	shard        *Shard
	globalSeqNo  uint64 // Use atomic.AddUint64, atomic.LoadUint64
	// Channels
	ReadBatch []*ReadRequest
	ReadChan  chan *ReadRequest
}

func NewCentralized(name string, globalConfig common.GlobalConfig, follower common.FollowerInterface, isLeader bool) *Centralized {
	c := &Centralized{}
	c.log = log.New(os.Stdout, "["+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	c.name = name
	c.follower = follower
	c.isLeader = isLeader

	c.globalConfig.Store(globalConfig)
	c.shard = NewShard(name, "../pird/pir.socket", globalConfig)
	c.globalSeqNo = 0
	c.ReadBatch = make([]*ReadRequest, 0)
	c.ReadChan = make(chan *ReadRequest)
	go c.batchReads()

	return c
}

/** PUBLIC METHODS (threadsafe) **/
func (c *Centralized) GetName() string {
	return c.name
}

func (c *Centralized) Ping(args *common.PingArgs, reply *common.PingReply) error {
	c.log.Println("Ping: " + args.Msg + ", ... Pong")

	// Try to ping the follower if one exists
	if c.follower != nil {
		var fReply common.PingReply
		fErr := c.follower.Ping(&common.PingArgs{"PING"}, &fReply)
		if fErr != nil {
			reply.Err = c.follower.GetName() + " Ping failed"
		} else {
			reply.Err = fReply.Err
		}
	} else {
		reply.Err = ""
	}

	reply.Msg = "PONG"
	return nil
}

func (c *Centralized) Write(args *common.WriteArgs, reply *common.WriteReply) error {
	//c.log.Println("Write: ")

	// @todo --- parallelize writes.
	if c.isLeader {
		seqNo := atomic.AddUint64(&c.globalSeqNo, 1)
		args.GlobalSeqNo = seqNo
	}

	c.shard.Write(args, &common.WriteReply{})
	if c.follower != nil {
		var fReply common.WriteReply
		fErr := c.follower.Write(args, &fReply)
		if fErr != nil {
			// Assume all servers always available
			reply.Err = fErr.Error()
			c.log.Fatalf("Error forwarding to follower %v", c.follower.GetName())
			return fErr
		} else {
			reply.Err = fReply.Err
		}
	} else {
		reply.Err = ""
	}

	return nil
}

func (c *Centralized) Read(args *common.ReadArgs, reply *common.ReadReply) error {
	c.log.Println("Read: ")
	resultChan := make(chan []byte)
	c.ReadChan <- &ReadRequest{args, resultChan}
	reply.Err = ""
	reply.Data = <-resultChan
	return nil
}

func (c *Centralized) BatchRead(args *common.BatchReadArgs, reply *common.BatchReadReply) error {
	// Start local computation
	var fReply common.BatchReadReply
	myReplyChan := make(chan *common.BatchReadReply)
	c.shard.BatchRead(args, myReplyChan)

	// Send to followers
	if c.follower != nil {
		fErr := c.follower.BatchRead(args, &fReply)
		if fErr != nil {
			// Assume all servers always available
			reply.Err = fErr.Error()
			c.log.Fatalf("Error forwarding to follower %v", c.follower.GetName())
			return fErr
		} else {
			reply.Err = fReply.Err
		}
	} else {
		reply.Err = ""
	}

	// Combine results
	myReply := <-myReplyChan
	if c.follower != nil {
		for i, _ := range myReply.Replies {
			_ = myReply.Replies[i].Combine(fReply.Replies[i].Data)
		}
	}

	reply.Replies = myReply.Replies
	return nil
}

func (c *Centralized) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	c.log.Println("GetUpdates: ")
	// @TODO
	return nil
}

/** PRIVATE METHODS (singlethreaded) **/
func (c *Centralized) batchReads() {
	var readReq *ReadRequest
	for {
		select {
		case readReq = <-c.ReadChan:
			c.ReadBatch = append(c.ReadBatch, readReq)
			if len(c.ReadBatch) >= BATCH_SIZE {
				go c.triggerBatchRead(c.ReadBatch)
				c.ReadBatch = make([]*ReadRequest, 0)
			} else {
				c.log.Printf("Read: add to batch, size=%v\n", len(c.ReadBatch))
			}
		}
	}
}

func (c *Centralized) triggerBatchRead(batch []*ReadRequest) {
	args := &common.BatchReadArgs{}
	// Copy args
	args.Args = make([]common.ReadArgs, len(batch), len(batch))
	for i, val := range batch {
		args.Args[i] = *val.Args
	}

	// Choose a SeqNoRange
	currSeqNo := atomic.LoadUint64(&c.globalSeqNo) + 1
	globalConfig := c.globalConfig.Load().(common.GlobalConfig)
	args.SeqNoRange = common.Range{}
	args.SeqNoRange.Start = currSeqNo - uint64(globalConfig.WindowSize) // Inclusive
	if args.SeqNoRange.Start < 1 {
		args.SeqNoRange.Start = 1 // Minimum of 1
	}
	args.SeqNoRange.End = currSeqNo // Exclusive
	args.SeqNoRange.Aborted = make([]uint64, 0, 0)

	// Choose a RandSeed
	args.RandSeed = 0

	// Start computation on local shard
	var reply common.BatchReadReply
	err := c.BatchRead(args, &reply)
	if err != nil || reply.Err != "" {
		c.log.Fatalf("Error doing BatchRead: err=%v, replyErr=%v\n", err, reply.Err)
		return
	}

	// Respond to clients
	for i, val := range reply.Replies {
		batch[i].Reply(val.Data)
	}

}
