package libpdb

import (
	"github.com/ryscheng/pdb/common"
	"log"
	"os"
)

/**
 * Client interface for libpdb
 * Goroutines:
 * - 1x RequestManager.writePeriodic
 * - 1x RequestManager.readPeriodic
 */
type Client struct {
	log          *log.Logger
	name         string
	serverConfig *common.TrustDomainConfig
	serverRef    *common.TrustDomainRef
	msgReqMan    *RequestManager
}

func NewClient(name string, serverConfig *common.TrustDomainConfig) *Client {
	c := &Client{}
	c.log = log.New(os.Stdout, "[Client:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	c.name = name
	c.serverConfig = serverConfig
	c.serverRef = common.NewTrustDomainRef(name, serverConfig)

	// @todo
	c.msgReqMan = NewRequestManager(name, 8, c.serverRef)

	c.log.Println("NewClient: starting new client - " + name)
	return c
}

func (c *Client) Ping() bool {
	err, reply := c.serverRef.Ping()
	if err == nil && reply.Err == "" {
		return true
	} else {
		return false
	}
}
