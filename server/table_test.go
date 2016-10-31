package server

import (
	"github.com/ryscheng/pdb/pir"
	"log"
)

import "testing"

func TestTableSanity(t *testing.T) {
	status := make(chan int)
	go pir.CreateMockServer(status, "pir.socket")
	<-status

	pirServer, err := pir.Connect("pir.socket")
	if err != nil {
		t.Error(err)
		return
	}

	log := log.New(os.Stdout, "[testlog] ", log.Ldate|log.Ltime|log.Lshortfile)

	table := NewTable(pirServer, log, 4, 0.75, 0.01)

	table.Close()
	pirServer.Disconnect()

	status <- 1
}
