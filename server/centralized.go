package server

import (
	"github.com/ryscheng/pdb/common"
	"log"
	"os"
)

type Centralized struct {
	// Private State
	log             *log.Logger
	name            string
	dataLayerConfig *DataLayerConfig
	followerConfig  *common.TrustDomainConfig
	isLeader        bool

	followerRef *common.TrustDomainRef
	globalSeqNo uint64
	shard       *Shard
}

func NewCentralized(name string, followerConfig *common.TrustDomainConfig, isLeader bool) *Centralized {
	c := &Centralized{}
	c.log = log.New(os.Stdout, "[Frontend:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	c.name = name
	c.followerConfig = followerConfig
	c.isLeader = isLeader

	c.followerRef = common.NewTrustDomainRef(name, followerConfig)
	c.globalSeqNo = 1
	c.shard = NewShard(name)

	return c
}

/** PUBLIC METHODS (threadsafe) **/
func (c *Centralized) Ping(args *common.PingArgs, reply *common.PingReply) error {
	c.log.Println("Ping: " + args.Msg + ", ... Pong")

	// Try to ping the follower if one exists
	fName, haveFollower := c.followerConfig.GetName()
	if haveFollower {
		fErr, fReply := c.followerRef.Ping()
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

	// @TODO - need to lock
	if c.isLeader {
		args.GlobalSeqNo = c.globalSeqNo
	}

	c.shard.Write(args, &common.WriteReply{})
	fName, haveFollower := c.followerConfig.GetName()
	if haveFollower {
		fErr, fReply := c.followerRef.Write(args)

	}

	// Only if successfully forwarded
	c.globalSeqNo += 1

	return nil
}

func (c *Centralized) Read(args *common.ReadArgs, reply *common.ReadReply) error {
	c.log.Println("Read: ")
	// @TODO
	return nil
}

func (c *Centralized) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	c.log.Println("GetUpdates: ")
	// @TODO
	return nil
}
