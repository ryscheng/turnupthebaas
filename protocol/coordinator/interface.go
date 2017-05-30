package coordinator

import (
	"github.com/privacylab/talek/common"
)

// Interface is the interface to the central coordinator
type Interface interface {
	GetInfo(args *interface{}, reply *GetInfoReply) error
	GetCommonConfig(args *interface{}, reply *common.Config) error
	GetLayout(args *GetLayoutArgs, reply *GetLayoutReply) error
	GetIntVec(args *GetIntVecArgs, reply *GetIntVecReply) error
	Commit(args *CommitArgs, reply *CommitReply) error
}
