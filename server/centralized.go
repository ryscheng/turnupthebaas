package server

import (
	"github.com/ryscheng/pdb/common"
	"golang.org/x/net/trace"
	"sync/atomic"
)

type Centralized struct {
	/** Private State **/
	// Static
	log      *common.Logger
	name     string
	follower common.FollowerInterface
	isLeader bool
	status   chan int

	// Thread-safe
	config         atomic.Value //ServerConfig
	shard          *Shard
	proposedSeqNo  uint64 // Use atomic.AddUint64, atomic.LoadUint64
	committedSeqNo uint64 // Use atomic.AddUint64, atomic.LoadUint64
	// Channels
	ReadBatch []*common.ReadRequest
	ReadChan  chan *common.ReadRequest
	closeChan chan int
}

func NewCentralized(name string, socket string, config ServerConfig, follower common.FollowerInterface, isLeader bool) *Centralized {
	c := &Centralized{}
	c.log = common.NewLogger(name)
	c.name = name
	c.follower = follower
	c.isLeader = isLeader

	c.config.Store(config)

	c.closeChan = make(chan int)
	c.shard = NewShard(name, socket, config)

	c.proposedSeqNo = 0
	c.committedSeqNo = 0
	c.ReadBatch = make([]*common.ReadRequest, 0)
	c.ReadChan = make(chan *common.ReadRequest)
	go c.batchReads()

	return c
}

func (c *Centralized) Close() {
	// stop processing.
	c.closeChan <- 1
	// Stop the shard.
	c.shard.Close()
}

/** PUBLIC METHODS (threadsafe) **/
func (c *Centralized) GetName() string {
	return c.name
}

func (c *Centralized) Ping(args *common.PingArgs, reply *common.PingReply) error {
	c.log.Trace.Println("Ping: enter")
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
	c.log.Trace.Println("Ping: exit")
	c.log.Info.Println("Ping: " + args.Msg + ", ... Pong")
	return nil
}

func (c *Centralized) Write(args *common.WriteArgs, reply *common.WriteReply) error {
	c.log.Trace.Println("Write: enter")
	tr := trace.New("centralized.write", "Write")
	defer tr.Finish()

	// @todo --- parallelize writes.
	if c.isLeader {
		seqNo := atomic.AddUint64(&c.proposedSeqNo, 1)
		args.GlobalSeqNo = seqNo
	}

	c.shard.Write(args)
	if c.follower != nil {
		var fReply common.WriteReply
		fErr := c.follower.Write(args, &fReply)
		if fErr != nil {
			// Assume all servers always available
			reply.Err = fErr.Error()
			c.log.Error.Fatalf("Error forwarding to follower %v", c.follower.GetName())
			return fErr
		} else {
			reply.Err = fReply.Err
		}
	} else {
		reply.Err = ""
	}

	atomic.StoreUint64(&c.committedSeqNo, args.GlobalSeqNo)
	reply.GlobalSeqNo = args.GlobalSeqNo
	c.log.Trace.Println("Write: exit")
	return nil
}

func (c *Centralized) Read(args *common.ReadArgs, reply *common.ReadReply) error {
	c.log.Trace.Println("Read: enter")
	tr := trace.New("centralized.read", "Read")
	defer tr.Finish()
	resultChan := make(chan *common.ReadReply)
	c.ReadChan <- &common.ReadRequest{args, resultChan}
	myReply := <-resultChan
	*reply = *myReply
	c.log.Trace.Println("Read: exit")
	return nil
}

func (c *Centralized) BatchRead(args *common.BatchReadRequest, reply *common.BatchReadReply) error {
	c.log.Trace.Println("BatchRead: enter")
	tr := trace.New("centralized.batchread", "BatchRead")
	defer tr.Finish()
	// Start local computation
	var fReply common.BatchReadReply
	myReplyChan := make(chan *common.BatchReadReply)
	args.ReplyChan = myReplyChan
	c.shard.BatchRead(args)

	// Send to followers
	if c.follower != nil {
		fErr := c.follower.BatchRead(args, &fReply)
		if fErr != nil {
			// Assume all servers always available
			reply.Err = fErr.Error()
			c.log.Error.Fatalf("Error forwarding to follower %v", c.follower.GetName())
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
	c.log.Trace.Println("BatchRead: exit")
	return nil
}

func (c *Centralized) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	c.log.Trace.Println("GetUpdates: ")
	// @TODO
	return nil
}

/** PRIVATE METHODS (singlethreaded) **/
func (c *Centralized) batchReads() {
	config := c.config.Load().(ServerConfig)
	var readReq *common.ReadRequest
	for {
		select {
		case readReq = <-c.ReadChan:
			c.ReadBatch = append(c.ReadBatch, readReq)
			if len(c.ReadBatch) >= config.ReadBatch {
				go c.triggerBatchRead(c.ReadBatch)
				c.ReadBatch = make([]*common.ReadRequest, 0, config.ReadBatch)
			} else {
				c.log.Trace.Printf("Read: add to batch, size=%v\n", len(c.ReadBatch))
			}
			continue
		case <-c.closeChan:
			break
		}
	}
}

func (c *Centralized) triggerBatchRead(batch []*common.ReadRequest) {
	c.log.Trace.Println("triggerBatchRead: enter")
	args := &common.BatchReadRequest{}
	// Copy args
	args.Args = make([]common.PirArgs, len(batch), len(batch))
	for i, val := range batch {
		args.Args[i] = val.Args.ForTd[0]
	}

	// Choose a SeqNoRange
	currSeqNo := atomic.LoadUint64(&c.committedSeqNo) + 1
	config := c.config.Load().(ServerConfig)
	args.SeqNoRange = common.Range{}
	if currSeqNo <= uint64(config.CommonConfig.WindowSize()) {
		args.SeqNoRange.Start = 1 // Minimum of 1
	} else {
		args.SeqNoRange.Start = currSeqNo - uint64(config.CommonConfig.WindowSize()) // Inclusive
	}
	args.SeqNoRange.End = currSeqNo // Exclusive
	args.SeqNoRange.Aborted = make([]uint64, 0, 0)

	// Choose a RandSeed
	args.RandSeed = 0

	// Start computation on local shard
	var reply common.BatchReadReply
	err := c.BatchRead(args, &reply)
	if err != nil || reply.Err != "" {
		c.log.Error.Fatalf("Error doing BatchRead: err=%v, replyErr=%v\n", err, reply.Err)
		return
	}

	// logging of throughput.

	// Respond to clients
	for i, val := range reply.Replies {
		val.GlobalSeqNo = args.SeqNoRange
		batch[i].Reply(&val)
	}

	c.log.Trace.Println("triggerBatchRead: exit")
}
