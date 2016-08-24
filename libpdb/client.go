package libpdb

import (
	"github.com/ryscheng/pdb/common"
	"log"
	"os"
	"sync/atomic"
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
	globalConfig atomic.Value //common.GlobalConfig
	leaderRef    *common.TrustDomainRef
	msgReqMan    *RequestManager
}

func NewClient(name string, globalConfig common.GlobalConfig) *Client {
	c := &Client{}
	c.log = log.New(os.Stdout, "[Client:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	c.name = name
	c.globalConfig.Store(globalConfig)
	c.leaderRef = common.NewTrustDomainRef(name, globalConfig.TrustDomains[0])

	c.msgReqMan = NewRequestManager(name, c.leaderRef, &c.globalConfig)

	c.log.Println("NewClient: starting new client - " + name)
	return c
}

/** PUBLIC METHODS (threadsafe) **/
func (c *Client) Ping() bool {
	err, reply := c.leaderRef.Ping()
	if err == nil && reply.Err == "" {
		return true
	} else {
		return false
	}
}

func (c *Client) SetGlobalConfig(globalConfig common.GlobalConfig) {
	c.globalConfig.Store(globalConfig)
}
