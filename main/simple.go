package main

import (
	"github.com/ryscheng/pdb/libpdb"
	"github.com/ryscheng/pdb/server"
	"log"
	"time"
)

func main() {
	log.Println("Simple Sanity Test")
	s := server.NewShardServer(9000, 0)
	c := libpdb.NewClient("c1", []string{"localhost:9000"})
	c.Ping()
	time.Sleep(10 * time.Second)
	s.Kill()
}
