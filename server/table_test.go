package server

import (
	"bytes"
	"github.com/ryscheng/pdb/cuckoo"
	"github.com/ryscheng/pdb/pir"
	"log"
	"os"
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

	pirServer.Configure(128, 128, 128)

	log := log.New(os.Stdout, "[testlog] ", log.Ldate|log.Ltime|log.Lshortfile)

	table := NewTable(pirServer, "testtable", log, 4, 0.75, 0.01)

	table.Close()
	pirServer.Disconnect()

	status <- 1
}

func TestTableWritingConsistency(t *testing.T) {
	status := make(chan int)
	go pir.CreateMockServer(status, "pir.socket")
	<-status

	pirServer, err := pir.Connect("pir.socket")
	if err != nil {
		t.Error(err)
		return
	}

	pirServer.Configure(128, 128, 128)

	log := log.New(os.Stdout, "[testlog] ", log.Ldate|log.Ltime|log.Lshortfile)

	table := NewTable(pirServer, "testtable", log, 4, 0.75, 0.01)
	data := bytes.NewBufferString("Test Data").Bytes()

	for i := 1; i < 128; i += 1 {
		err = table.Write(&cuckoo.Item{data, 0, i})
		if err != nil {
			t.Error(err)
			return
		}
	}
	err = table.Flop()
	if err != nil {
		t.Error(err)
		return
	}

	table.Close()
	pirServer.Disconnect()

	status <- 1
}
