package server

import (
	"github.com/ryscheng/pdb/common"
	"log"
	"os"
)

type Frontend struct {
	// Private State
	log             *log.Logger
	name            string
	dataLayerConfig *DataLayerConfig
	followerConfig  *common.TrustDomainConfig
	isLeader        bool

	//dataLayerRef *DataLayerRef
	followerRef *common.TrustDomainRef
}

func NewFrontend(name string, dataLayerConfig *DataLayerConfig, followerConfig *common.TrustDomainConfig, isLeader bool) *Frontend {
	fe := &Frontend{}
	fe.log = log.New(os.Stdout, "[Frontend:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	fe.name = name
	fe.dataLayerConfig = dataLayerConfig
	fe.followerConfig = followerConfig
	fe.isLeader = isLeader

	fe.followerRef = common.NewTrustDomainRef(name, followerConfig)

	return fe
}

/** PUBLIC METHODS (threadsafe) **/
func (fe *Frontend) Ping(args *common.PingArgs, reply *common.PingReply) error {
	fe.log.Println("Ping: " + args.Msg + ", ... Pong")

	// Try to ping the follower if one exists
	fName, haveFollower := fe.followerConfig.GetName()
	if haveFollower {
		fErr, fReply := fe.followerRef.Ping()
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

func (fe *Frontend) Write(args *common.WriteArgs, reply *common.WriteReply) error {
	fe.log.Println("Write: ")
	// @TODO
	return nil
}

func (fe *Frontend) Read(args *common.ReadArgs, reply *common.ReadReply) error {
	fe.log.Println("Read: ")
	// @TODO
	return nil
}

func (fe *Frontend) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	fe.log.Println("GetUpdates: ")
	// @TODO
	return nil
}
