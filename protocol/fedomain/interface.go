package fedomain

import (
	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/protocol/layout"
	"github.com/privacylab/talek/protocol/notify"
)

// Interface is the interface to a trust domain frontend
type Interface interface {
	GetInfo(args *interface{}, reply *GetInfoReply) error
	GetLayout(args *layout.GetLayoutArgs, reply *layout.GetLayoutReply) error
	Notify(args *notify.Args, reply *notify.Reply) error
	Write(args *common.WriteArgs, reply *common.WriteReply) error
	EncPIR(args *EncPIRArgs, reply *EncPIRReply) error
}
