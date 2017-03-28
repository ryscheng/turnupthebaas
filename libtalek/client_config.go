package libtalek

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/privacylab/talek/common"
)

// ClientConfig represents the configuration parameters to Talek needed by
// the client.
type ClientConfig struct {
	*common.CommonConfig `json:"-"`

	// How often should Writes be made to the server
	WriteInterval time.Duration `json:",string"`

	// How often should reads be made to the server
	ReadInterval time.Duration `json:",string"`

	// Where are the different servers?
	TrustDomains []*common.TrustDomainConfig
}

// ClientConfigFromFile restores a client configuration from on-disk form.
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
