package main

import (
	"github.com/ryscheng/pdb/libpdb"
	"github.com/ryscheng/pdb/server"
	"log"
)

func main() {
	log.Println("Simple Sanity Test")
	go server.NewFrontend(9000)
	c := libpdb.NewClient("c1", []string{"http://localhost:9000"})
	c.Ping()
}
