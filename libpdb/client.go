package libpdb

import (
	"github.com/ryscheng/pdb/common"
	"log"
	"net/rpc"
	"os"
)

type Client struct {
	log     *log.Logger
	name    string
	servers []string
}

func NewClient(name string, servers []string) *Client {
	c := &Client{}
	c.log = log.New(os.Stdout, "[client:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	c.name = name
	c.servers = servers
	c.log.Println("NewClient: starting new client - " + name)
	return c
}

func (c *Client) Ping() bool {
	c.log.Printf("Ping: enter\n")
	client, err := rpc.Dial("tcp", c.servers[0])
	if err != nil {
		log.Fatal("dialing:", err)
	}
	args := &common.PingArgs{"PING"}
	var reply common.PingReply
	err = client.Call("FrontEndRpc.Ping", args, &reply)
	if err != nil {
		log.Fatal("rpc error:", err)
	}
	client.Close()
	c.log.Printf("Ping: %v, %v", args, reply)
	return true
}
