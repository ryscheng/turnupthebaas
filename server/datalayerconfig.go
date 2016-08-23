package server

type DataLayerConfig struct {
	ServerAddr map[string]map[string]string //groupName -> serverName -> serverAddr
}
