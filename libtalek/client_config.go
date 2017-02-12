package libtalek

import (
	"encoding/json"
	"github.com/privacylab/talek/common"
	"io/ioutil"
	"time"
)

type ClientConfig struct {
	*common.CommonConfig `json:"-"`

	// How often should Writes be made to the server
	WriteInterval time.Duration

	// How often should reads be made to the server
	ReadInterval time.Duration

	// Where are the different servers?
	TrustDomains []*common.TrustDomainConfig `json:"-"`
}

// Load configuration from a JSON file. returns the config on success or nil if
// loading or parsing the file fails.
func ClientConfigFromFile(file string, commonBase *common.CommonConfig) *ClientConfig {
	configString, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}
	config := new(ClientConfig)
	if err := json.Unmarshal(configString, config); err != nil {
		return nil
	}
	config.CommonConfig = commonBase
	return config
}
