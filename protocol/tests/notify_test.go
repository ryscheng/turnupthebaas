package tests

import (
	"testing"

	protocol "github.com/privacylab/talek/protocol/notify"
	server "github.com/privacylab/talek/server/replica"
)

func TestNotify(t *testing.T) {
	testAddr := randAddr()
	s, err := server.NewServer("test", testAddr, true, testConfig())
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
