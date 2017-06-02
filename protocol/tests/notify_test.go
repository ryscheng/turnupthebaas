package tests

import (
	"testing"

	protocol "github.com/privacylab/talek/protocol/notify"
	"github.com/privacylab/talek/server/fedomain"
	"github.com/privacylab/talek/server/feglobal"
	"github.com/privacylab/talek/server/replica"
)

func TestNotifyReplica(t *testing.T) {
	testAddr := randAddr()
	s, err := replica.NewServer("test", testAddr, true, testConfig(), 0, "cpu.0")
	if err != nil {
		t.Errorf("Error creating new server")
	}
	c := protocol.NewClient("test", testAddr)
	var cc protocol.Interface
	cc = c

	if err = cc.Notify(&protocol.Args{}, &protocol.Reply{}); err != nil {
		t.Errorf("Error calling Notify: %v", err)
	}

	c.Close()
	s.Close()
}

func TestNotifyFEGlobal(t *testing.T) {
	testAddr := randAddr()
	s, err := feglobal.NewServer("test", testAddr, true, testConfig())
	if err != nil {
		t.Errorf("Error creating new server")
	}
	c := protocol.NewClient("test", testAddr)
	var cc protocol.Interface
	cc = c

	if err = cc.Notify(&protocol.Args{}, &protocol.Reply{}); err != nil {
		t.Errorf("Error calling Notify: %v", err)
	}

	c.Close()
	s.Close()
}

func TestNotifyFEDomain(t *testing.T) {
	testAddr := randAddr()
	s, err := fedomain.NewServer("test", testAddr, true, testConfig())
	if err != nil {
		t.Errorf("Error creating new server")
	}
	c := protocol.NewClient("test", testAddr)
	var cc protocol.Interface
	cc = c

	if err = cc.Notify(&protocol.Args{}, &protocol.Reply{}); err != nil {
		t.Errorf("Error calling Notify: %v", err)
	}

	c.Close()
	s.Close()
}
