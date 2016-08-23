package server

import (
	"github.com/ryscheng/pdb/common"
	"log"
	"os"
)

type Frontend struct {
	// Private State
	log  *log.Logger
	name string
}

func NewFrontend(name string) *Frontend {
	fe := &Frontend{}
	fe.log = log.New(os.Stdout, "[Frontend:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	fe.name = name

	return fe
}

/** PUBLIC METHODS (threadsafe) **/
func (fe *Frontend) Ping(args *common.PingArgs, reply *common.PingReply) error {
	fe.log.Println("Ping: " + args.Msg + ", ... Pong")
	var err error = nil
	reply.Err = ""
	reply.Msg = "PONG"
	return err
}

func (fe *Frontend) Write(args *common.WriteArgs, reply *common.WriteReply) error {
	fe.log.Println("Write: ")
	var err error = nil
	// @TODO
	reply.Err = ""
	return err
}

func (fe *Frontend) Read(args *common.ReadArgs, reply *common.ReadReply) error {
	fe.log.Println("Read: ")
	var err error = nil
	// @TODO
	reply.Err = ""
	return err
}

func (fe *Frontend) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	fe.log.Println("GetUpdates: ")
	var err error = nil
	// @TODO
	reply.Err = ""
	return err
}
