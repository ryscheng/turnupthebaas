package libpdb

import (
	"log"
	"net/http"
	"os"
)

type RequestManager struct {
	log     *log.Logger
	name    string
	servers []string
}

func NewRequestManager(name string, servers []string) *RequestManager {
	rm := &RequestManager{}
	rm.log = log.New(os.Stdout, "[RequestManager:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	rm.name = name
	rm.servers = servers
	rm.log.Println("NewRequestManager: starting new RequestManager - " + name)
	return rm
}

func (rm *RequestManager) Ping() bool {
	resp, err := http.Get(rm.servers[0])
	log.Println(resp)
	log.Println(err)
	return true
}
