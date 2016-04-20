package main

import (
	"github.com/ryscheng/pdb/server"
	"log"
)

func main() {
	log.Println("------------------")
	log.Println("--- PDB Server ---")
	log.Println("------------------")
	fe := server.New(9999)
	log.Println(fe)
}
