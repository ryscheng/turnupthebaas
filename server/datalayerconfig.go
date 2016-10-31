package server

import (
	"github.com/ryscheng/pdb/common"
)

type DataLayerConfig struct {
	*common.GlobalConfig
	ServerAddr map[string]map[string]string //groupName -> serverName -> serverAddr
}
