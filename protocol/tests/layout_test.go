package tests

import (
	"testing"
	"time"

	protocol "github.com/privacylab/talek/protocol/layout"
	"github.com/privacylab/talek/server/coordinator"
	"github.com/privacylab/talek/server/fedomain"
)

func TestLayoutCoordinator(t *testing.T) {
	testAddr := randAddr()
	s, err := coordinator.NewServer("test", testAddr, true, testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	c := protocol.NewClient("test", testAddr)
	var cc protocol.Interface
	cc = c

	if err = cc.GetLayout(&protocol.Args{}, &protocol.Reply{}); err != nil {
		t.Errorf("Error calling GetLayout: %v", err)
	}

	c.Close()
	s.Close()
}

func TestLayoutFEDomain(t *testing.T) {
	testAddr := randAddr()
	s, err := fedomain.NewServer("test", testAddr, true, testConfig())
	if err != nil {
		t.Errorf("Error creating new server")
	}
	c := protocol.NewClient("test", testAddr)
	var cc protocol.Interface
	cc = c

	if err = cc.GetLayout(&protocol.Args{}, &protocol.Reply{}); err != nil {
		t.Errorf("Error calling GetLayout: %v", err)
	}

	c.Close()
	s.Close()
}
