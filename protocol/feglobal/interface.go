package feglobal

import (
	"github.com/privacylab/talek/protocol/intvec"
	"github.com/privacylab/talek/protocol/notify"
)

// Interface is the interface to the global frontend
type Interface interface {
	GetInfo(args *interface{}, reply *GetInfoReply) error
	GetIntVec(args *intvec.Args, reply *intvec.Reply) error
	Notify(args *notify.Args, reply *notify.Reply) error
	Write(args *WriteArgs, reply *WriteReply) error
	Read(args *ReadArgs, reply *ReadReply) error
}
