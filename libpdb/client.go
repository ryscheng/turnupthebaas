package libpdb

import (
	"log"
	"os"
)

type Client struct {
	log        *log.Logger
	name       string
	servers    []string
	requestMan *RequestManager
}

func NewClient(name string, servers []string) *Client {
	c := &Client{}
	c.log = log.New(os.Stdout, "[Client:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	c.name = name
	c.servers = servers
	c.requestMan = NewRequestManager("1", servers)
	c.log.Println("NewClient: starting new client - " + name)
	return c
}

func (c *Client) Ping() bool {
	c.requestMan.Ping()
	return true
}
