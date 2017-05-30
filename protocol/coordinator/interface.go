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
	GetLayout(args *layout.GetLayoutArgs, reply *layout.GetLayoutReply) error
	GetIntVec(args *intvec.GetIntVecArgs, reply *intvec.GetIntVecReply) error
	Commit(args *CommitArgs, reply *CommitReply) error
}
