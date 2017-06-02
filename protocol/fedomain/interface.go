package fedomain

import (
	"github.com/privacylab/talek/protocol/feglobal"
	"github.com/privacylab/talek/protocol/layout"
	"github.com/privacylab/talek/protocol/notify"
)

// Interface is the interface to a trust domain frontend
type Interface interface {
	GetInfo(args *interface{}, reply *GetInfoReply) error
	GetLayout(args *layout.Args, reply *layout.Reply) error
	Notify(args *notify.Args, reply *notify.Reply) error
	Write(args *feglobal.WriteArgs, reply *feglobal.WriteReply) error
	EncPIR(args *feglobal.EncPIRArgs, reply *feglobal.ReadReply) error
}
