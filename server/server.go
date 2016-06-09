package server

import (
	"log"
)

type Server struct {
	log *log.Logger
}

func NewServer() *Server {
	s := &Server{}
	s.log = log.New(os.Stdout, "[server] ", log.Ldate|log.Ltime|log.Lshortfile)
	return s
}
