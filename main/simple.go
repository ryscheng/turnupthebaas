package main

import (
	"github.com/ryscheng/pdb/libpdb"
	"github.com/ryscheng/pdb/server"
	"log"
)

func main() {
	log.Println("Simple Sanity Test")
	s := server.NewServer(9000, 0)
	c := libpdb.NewClient("c1", []string{"localhost:9000"})
	c.Ping()
	s.Kill()
}
