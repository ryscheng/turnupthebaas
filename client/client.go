package client

import (
	"log"
	"net/http"
	"os"
)

type Client struct {
	log     *log.Logger
	name    string
	servers []string
}

func New(name string, servers []string) *Client {
	c := &Client{}
	c.log = log.New(os.Stdout, "[client] ", log.Ldate|log.Ltime|log.Lshortfile)
	c.name = name
	c.log.Println("New: starting new client - " + name)
	return c
}

func (c *Client) Ping() bool {

	return true
}
