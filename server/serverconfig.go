package server

import (
	"github.com/ryscheng/pdb/common"
)

type ServerConfig struct {
	*common.CommonConfig

	// The names of the different servers participating as leader/followers within
	// a single trust domain
	ServerAddrs map[string]map[string]string //groupName -> serverName -> serverAddr
}
