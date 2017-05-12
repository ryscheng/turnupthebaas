package coordinator

import (
	"github.com/privacylab/talek/common"
)

// Interface is the interface to the central coordinator
type Interface interface {
	GetInfo(args *interface{}, reply *GetInfoReply) error
	GetCommonConfig(args *interface{}, reply *common.Config) error
	Commit(args *CommitArgs, reply *CommitReply) error
	GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error
}
