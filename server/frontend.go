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
	follower        common.FollowerInterface
	isLeader        bool

	//dataLayerRef *DataLayerRef
}

func NewFrontend(name string, dataLayerConfig *DataLayerConfig, follower common.FollowerInterface, isLeader bool) *Frontend {
	fe := &Frontend{}
	fe.log = log.New(os.Stdout, "[Frontend:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	fe.name = name
	fe.dataLayerConfig = dataLayerConfig
	fe.follower = follower
	fe.isLeader = isLeader

	return fe
}

/** PUBLIC METHODS (threadsafe) **/
func (fe *Frontend) Ping(args *common.PingArgs, reply *common.PingReply) error {
	fe.log.Println("Ping: " + args.Msg + ", ... Pong")

	// Try to ping the follower if one exists
	if fe.follower != nil {
		var fReply common.PingReply
		fErr := fe.follower.Ping(&common.PingArgs{"PING"}, &fReply)
		if fErr != nil {
			reply.Err = fe.follower.GetName() + " Ping failed"
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
