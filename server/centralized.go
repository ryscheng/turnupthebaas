package server

import (
	"github.com/ryscheng/pdb/common"
	"log"
	"os"
	"sync/atomic"
)

const BATCH_SIZE = 1

type Centralized struct {
	// Private State
	log             *log.Logger
	name            string
	dataLayerConfig *DataLayerConfig
	follower        common.FollowerInterface
	isLeader        bool

	// Thread-safe
	shard       *Shard
	globalSeqNo uint64 // Use atomic.AddUint64
	// Channels
	ReadBatch []*ReadRequest
	ReadChan  chan *ReadRequest
}

func NewCentralized(name string, follower common.FollowerInterface, isLeader bool) *Centralized {
	c := &Centralized{}
	c.log = log.New(os.Stdout, "["+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	c.name = name
	c.follower = follower
	c.isLeader = isLeader

	c.shard = NewShard(name)
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
	return nil
}

func (c *Centralized) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	c.log.Println("GetUpdates: ")
	// @TODO
	return nil
}

func (c *Centralized) batchReads() {
	var readReq *ReadRequest
	for {
		select {
		case readReq = <-c.ReadChan:
			c.processRead(readReq)
		}
	}
}

func (c *Centralized) processRead(req *ReadRequest) {
	c.ReadBatch = append(c.ReadBatch, req)
	if len(c.ReadBatch) >= BATCH_SIZE {

		c.ReadBatch = make([]*ReadRequest, 0)
	} else {
		c.log.Printf("Read: add to batch, size=%v\n", len(c.ReadBatch))
	}
}
