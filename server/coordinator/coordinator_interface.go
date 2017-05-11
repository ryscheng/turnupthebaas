package coordinator

import (
	"github.com/privacylab/talek/common"
)

// CoordinatorInterface is the interface to the central coordinator
type CoordinatorInterface interface {
	GetInfo(args *interface{}, reply *GetInfoReply) error
	Commit(args *CommitArgs, reply *CommitReply) error
	GetConfig(args *interface{}, reply *common.Config) error
	GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error
}
