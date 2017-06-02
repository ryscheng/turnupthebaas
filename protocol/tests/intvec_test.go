package tests

import (
	"testing"
	"time"

	protocol "github.com/privacylab/talek/protocol/intvec"
	"github.com/privacylab/talek/server/coordinator"
	"github.com/privacylab/talek/server/feglobal"
)

func TestIntVecCoordinator(t *testing.T) {
	testAddr := randAddr()
	s, err := coordinator.NewServer("test", testAddr, true, testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	c := protocol.NewClient("test", testAddr)
	var cc protocol.Interface
	cc = c

	if err = cc.GetIntVec(&protocol.Args{}, &protocol.Reply{}); err != nil {
		t.Errorf("Error calling GetIntVec: %v", err)
	}

	c.Close()
	s.Close()
}

func TestIntVecFEGlobal(t *testing.T) {
	testAddr := randAddr()
	s, err := feglobal.NewServer("test", testAddr, true, testConfig())
	if err != nil {
		t.Errorf("Error creating new server")
	}
	c := protocol.NewClient("test", testAddr)
	var cc protocol.Interface
	cc = c

	if err = cc.GetIntVec(&protocol.Args{}, &protocol.Reply{}); err != nil {
		t.Errorf("Error calling GetIntVec: %v", err)
	}

	c.Close()
	s.Close()
}
