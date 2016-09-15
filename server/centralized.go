package server

import (
	"github.com/ryscheng/pdb/common"
	"log"
	"os"
	"sync"
)

type Centralized struct {
	// Private State
	log             *log.Logger
	name            string
	dataLayerConfig *DataLayerConfig
	followerConfig  *common.TrustDomainConfig
	isLeader        bool
	followerRef     *common.TrustDomainRef

	// Thread-safe
	shard *Shard

	// Unsafe
	mu          sync.Mutex
	globalSeqNo uint64
}

func NewCentralized(name string, followerConfig *common.TrustDomainConfig, isLeader bool) *Centralized {
	c := &Centralized{}
	c.log = log.New(os.Stdout, "[Frontend:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	c.name = name
	c.followerConfig = followerConfig
	c.isLeader = isLeader
	c.followerRef = common.NewTrustDomainRef(name, followerConfig)

	c.shard = NewShard(name)

	c.mu = sync.Mutex{}
	c.globalSeqNo = 1

	return c
}

/** PUBLIC METHODS (threadsafe) **/
func (c *Centralized) Ping(args *common.PingArgs, reply *common.PingReply) error {
	c.log.Println("Ping: " + args.Msg + ", ... Pong")

	// Try to ping the follower if one exists
	fName, haveFollower := c.followerConfig.GetName()
	if haveFollower {
		fReply, fErr := c.followerRef.Ping()
		if fErr != nil {
			reply.Err = fName + " Ping failed"
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
	c.log.Println("Write: ")

	// @todo --- parallelize writes.
	c.mu.Lock()
	if c.isLeader {
		args.GlobalSeqNo = c.globalSeqNo
	}

	c.shard.Write(args, &common.WriteReply{})
	fName, haveFollower := c.followerConfig.GetName()
	if haveFollower {
		fReply, fErr := c.followerRef.Write(args)
		if fErr != nil {
			// Assume all servers always available
			c.log.Fatalf("Error forwarding to follower %v", c.followerConfig)
			c.mu.Unlock()
			return fErr
		}
	}

	// Only if successfully forwarded
	c.globalSeqNo += 1

	c.mu.Unlock()
	return nil
}

func (c *Centralized) Read(args *common.ReadArgs, reply *common.ReadReply) error {
	c.log.Println("Read: ")
	c.shard.Read(args, reply)
	return nil
}

func (c *Centralized) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	c.log.Println("GetUpdates: ")
	// @TODO
	return nil
}
