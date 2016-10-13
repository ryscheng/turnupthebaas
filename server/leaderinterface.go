package server

import (
	"github.com/ryscheng/pdb/common"
)

type LeaderInterface interface {
	Ping(args *common.PingArgs, reply *common.PingReply) error
	Write(args *common.WriteArgs, reply *common.WriteReply) error
	Read(args *common.ReadArgs, reply *common.ReadReply) error
	GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error
}
