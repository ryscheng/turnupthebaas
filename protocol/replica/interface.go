package replica

import (
	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/protocol/notify"
)

// Interface is the interface to the central coordinator
type Interface interface {
	GetInfo(args *interface{}, reply *GetInfoReply) error
	Notify(args *notify.Args, reply *notify.Reply) error
	Write(args *common.WriteArgs, reply *common.WriteReply) error
	PIR(args *PIRArgs, reply *PIRReply) error
}
