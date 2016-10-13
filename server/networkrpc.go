package server

import (
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync/atomic"
)

/**
 * Registers RPC server with a specified handler
 * Goroutines:
 * - 1x n.listener.Accept() loop
 * - New goroutine for every new connection
 */
type NetworkRpc struct {
	log      *log.Logger
	dead     int32
	handler  LeaderInterface
	port     int
	server   *rpc.Server
	listener net.Listener
}

func NewNetworkRpc(handler LeaderInterface, port int) *NetworkRpc {
	n := &NetworkRpc{}
	n.log = log.New(os.Stdout, "[NetworkRpc] ", log.Ldate|log.Ltime|log.Lshortfile)
	n.dead = 0
	n.handler = handler
	n.port = port
	// Register RPC
	n.server = rpc.NewServer()
	n.server.Register(handler)
	//rpc.Register(handler)
	l, e := net.Listen("tcp", ":"+strconv.Itoa(port))
	if e != nil {
		n.log.Fatal("listen error:", e)
	}
	n.listener = l
	// Start services
	n.log.Println("NewNetworkRpc: starting new server on port " + strconv.Itoa(port))
	//@todo figure out how to support graceful HTTP shutdown.
	//      Maybe (https://github.com/braintree/manners) or (https://github.com/tylerb/graceful)
	/**
	rpc.HandleHTTP()
	n.log.Fatal(http.Serve(l, nil))
	**/
	go func() {
		for n.isDead() == false {
			conn, err := n.listener.Accept()
			if err != nil && n.isDead() == false {
				n.log.Printf("Accept: error %v\n", err.Error())
				continue
			} else if err == nil && n.isDead() == false {
				//go rpc.ServeConn(conn)
				go n.server.ServeConn(conn)
			} else if err == nil {
				conn.Close()
			}
		}
	}()
	return n
}

func (n *NetworkRpc) Kill() {
	atomic.StoreInt32(&n.dead, 1)
	if n.listener != nil {
		n.listener.Close()
	}
}

func (n *NetworkRpc) isDead() bool {
	return atomic.LoadInt32(&n.dead) != 0
}
