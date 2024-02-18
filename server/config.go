package server

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/privacylab/talek/common"
)

// Config represents the configuration needed to start a Talek server.
// configurations can be generated through util/talekutil
type Config struct {
	*common.Config

	// How many read requests should be made of the PIR server at a time?
	ReadBatch int
	// What's the minimum frequency when pending writes should be applied?
	WriteInterval time.Duration `json:",string"`

	// What's the minimum frequency when pending reads should be applied?
	ReadInterval time.Duration `json:",string"`

	// The trust domain this server is within. Includes keychain for the server.
	TrustDomain *common.TrustDomainConfig
	// In client read requests, which index is relevant for this server.
	TrustDomainIndex int
}

// ConfigFromFile restores a json cofig. returns the config on success or nil if
// loading or parsing the file fails.
func ConfigFromFile(file string, commonBase *common.Config) *Config {
	configString, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}
	config := new(Config)
	if err := json.Unmarshal(configString, config); err != nil {
		return nil
	}
	config.Config = commonBase
	return config
}
