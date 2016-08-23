package libpdb

import (
	"github.com/ryscheng/pdb/common"
	"log"
	"net/rpc"
	"os"
)

/**
 * Client interface for libpdb
 * Goroutines:
 * - 1x RequestManager.writePeriodic
 * - 1x RequestManager.readPeriodic
 */
type Client struct {
	log       *log.Logger
	name      string
	server    string
	msgReqMan *RequestManager
}

func Call(addr string, methodName string, args interface{}, reply interface{}) error {
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		log.Printf("rpc dialing failed: %v\n", err)
		return err
	}
	defer client.Close()

	err = client.Call(methodName, args, reply)
	if err != nil {
		log.Printf("rpc error:", err)
		return err
	}
	return nil
}

func NewClient(name string, server string) *Client {
	c := &Client{}
	c.log = log.New(os.Stdout, "[Client:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	c.name = name
	c.server = server
	// @todo update
	c.msgReqMan = NewRequestManager(name, 8)
	c.log.Println("NewClient: starting new client - " + name)
	return c
}

func (c *Client) Ping() bool {
	c.log.Printf("Ping: enter\n")
	args := &common.PingArgs{"PING"}
	var reply common.PingReply
	err := Call(c.server, "Shard.Ping", args, &reply)

	c.log.Printf("%s.Ping(): %v, %v\n", c.server, args, reply)
	if err == nil {
		return true
	} else {
		return false
	}
}
