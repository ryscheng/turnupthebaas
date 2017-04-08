package server

import (
	"log"
	"os"

	"github.com/privacylab/talek/common"
)

// Frontend terminates client connections to the leader server.
// It is the point of global serialization, and establishes sequence numbers.
type Frontend struct {
	// Private State
	log  *log.Logger
	name string
	*Config
	replicas []*common.FollowerInterface
}

// NewFrontend creates a new Frontend for a provided configuration.
func NewFrontend(name string, serverConfig *Config, replicas []*common.FollowerInterface) *Frontend {
	fe := &Frontend{}
	fe.log = log.New(os.Stdout, "[Frontend:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	fe.name = name
	fe.Config = serverConfig
	fe.replicas = replicas

	return fe
}

/** PUBLIC METHODS (threadsafe) **/

// GetName exports the name of the server.
func (fe *Frontend) GetName(args *interface{}, reply *string) error {
	*reply = fe.name
	return nil
}

// GetConfig returns the current common configuration from the server.
func (fe *Frontend) GetConfig(args *interface{}, reply *common.Config) error {
	config := fe.Config
	*reply = *config.Config
	return nil
}

func (fe *Frontend) Write(args *common.WriteArgs, reply *common.WriteReply) error {
	fe.log.Println("Write: ")
	// @TODO
	return nil
}

func (fe *Frontend) Read(args *common.EncodedReadArgs, reply *common.ReadReply) error {
	fe.log.Println("Read: ")
	// @TODO
	return nil
}

// GetUpdates provides the most recent global interest vector deltas.
func (fe *Frontend) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	fe.log.Println("GetUpdates: ")
	// @TODO
	return nil
}
