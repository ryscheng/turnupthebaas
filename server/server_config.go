package server

import (
	"encoding/json"
	"github.com/ryscheng/pdb/common"
	"io/ioutil"
	"time"
)

type ServerConfig struct {
	*common.CommonConfig `json:"-"`

	// How many read requests should be made of the PIR server at a time?
	ReadBatch int
	// What's the minimum frequency when pending writes should be applied?
	WriteInterval  time.Duration

	// WHat's the minimum frequency when pending reads should be applied?
	ReadInterval   time.Duration


	// The names of the different servers participating as leader/followers within
	// a single trust domain
	ServerAddrs map[string]map[string]string //groupName -> serverName -> serverAddr
}

// Load configuration from a JSON file. returns the config on success or nil if
// loading or parsing the file fails.
func ServerConfigFromFile(file string, commonBase *common.CommonConfig) *ServerConfig {
	configString, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}
	config := new(ServerConfig)
	if err := json.Unmarshal(configString, config); err != nil {
		return nil
	}
	config.CommonConfig = commonBase
	return config
}
