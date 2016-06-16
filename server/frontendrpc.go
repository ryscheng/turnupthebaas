package server

import (
	"github.com/ryscheng/pdb/common"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync/atomic"
)

type FrontEndRpc struct {
	log      *log.Logger
	dead     int32
	port     int
	listener net.Listener
}

func NewFrontEndRpc(port int) *FrontEndRpc {
	fe := &FrontEndRpc{}
	fe.log = log.New(os.Stdout, "[frontendrpc] ", log.Ldate|log.Ltime|log.Lshortfile)
	fe.dead = 0
	fe.port = port
	// Register RPC
	rpc.Register(fe)
	l, e := net.Listen("tcp", ":"+strconv.Itoa(port))
	if e != nil {
		fe.log.Fatal("listen error:", e)
	}
	fe.listener = l
	// Start services
	fe.log.Println("NewFrontEndRpc: starting new server on port " + strconv.Itoa(port))
	//@todo figure out how to support graceful HTTP shutdown.
	//      Maybe (https://github.com/braintree/manners) or (https://github.com/tylerb/graceful)
	/**
	rpc.HandleHTTP()
	fe.log.Fatal(http.Serve(l, nil))
	**/
	go func() {
		for fe.isDead() == false {
			conn, err := fe.listener.Accept()
			if err != nil && fe.isDead() == false {
				fe.log.Printf("Accept: error %v\n", err.Error())
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()
	return fe
}

func (fe *FrontEndRpc) Kill() {
	atomic.StoreInt32(&fe.dead, 1)
	if fe.listener != nil {
		fe.listener.Close()
	}
}

func (fe *FrontEndRpc) isDead() bool {
	return atomic.LoadInt32(&fe.dead) != 0
}

func (fe *FrontEndRpc) Ping(args *common.PingArgs, reply *common.PingReply) error {
	fe.log.Println("Ping: " + args.Msg + ", ... Pong")
	reply.Err = ""
	reply.Msg = "PONG"
	var err error = nil
	return err
}
