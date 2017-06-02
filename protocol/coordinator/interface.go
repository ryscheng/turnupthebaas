package coordinator

import (
	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/protocol/intvec"
	"github.com/privacylab/talek/protocol/layout"
)

// Interface is the interface to the central coordinator
type Interface interface {
	GetInfo(args *interface{}, reply *GetInfoReply) error
	GetCommonConfig(args *interface{}, reply *common.Config) error
	GetLayout(args *layout.Args, reply *layout.Reply) error
	GetIntVec(args *intvec.Args, reply *intvec.Reply) error
	Commit(args *CommitArgs, reply *CommitReply) error
}
