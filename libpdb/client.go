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
	globalConfig *common.GlobalConfig
	leaderRef    *common.TrustDomainRef
	msgReqMan    *RequestManager
}

func NewClient(name string, globalConfig *common.GlobalConfig) *Client {
	c := &Client{}
	c.log = log.New(os.Stdout, "[Client:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	c.name = name
	c.globalConfig = globalConfig
	c.leaderRef = common.NewTrustDomainRef(name, globalConfig.TrustDomains[0])

	c.msgReqMan = NewRequestManager(name, 8, c.serverRef, globalConfig)

	c.log.Println("NewClient: starting new client - " + name)
	return c
}

func (c *Client) Ping() bool {
	err, reply := c.leaderRef.Ping()
	if err == nil && reply.Err == "" {
		return true
	} else {
		return false
	}
}
