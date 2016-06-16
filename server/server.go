package server

import (
	"log"
	"os"
)

type Server struct {
	log    *log.Logger
	feRpc  *FrontEndRpc
	feHttp *FrontEndHttp
}

func NewServer(rpcPort int, httpPort int) *Server {
	s := &Server{}
	s.log = log.New(os.Stdout, "[Server] ", log.Ldate|log.Ltime|log.Lshortfile)
	if rpcPort != 0 {
		s.feRpc = NewFrontEndRpc(rpcPort)
	}
	if httpPort != 0 {
		s.feHttp = NewFrontEndHttp(httpPort)
	}
	return s
}

func (s *Server) Kill() {
	s.feRpc.Kill()
}
