package server

import (
	"log"
	"os"
)

type Shard struct {
	log *log.Logger
}

func NewShard() *Shard {
	s := &Shard{}
	s.log = log.New(os.Stdout, "[Shard] ", log.Ldate|log.Ltime|log.Lshortfile)
	return s
}
