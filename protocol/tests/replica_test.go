package tests

import (
	"testing"

	"github.com/privacylab/talek/protocol/feglobal"
	"github.com/privacylab/talek/protocol/notify"
	protocol "github.com/privacylab/talek/protocol/replica"
	server "github.com/privacylab/talek/server/replica"
)

func TestReplica(t *testing.T) {
	testAddr := randAddr()
	s, err := server.NewServer("test", testAddr, true, testConfig(), 0, "cpu.0")
	if err != nil {
		t.Errorf("Error creating new server")
	}
	c := protocol.NewClient("test", testAddr)
	var cc protocol.Interface
	cc = c

	if err = cc.GetInfo(nil, &protocol.GetInfoReply{}); err != nil {
		t.Errorf("Error calling GetInfo: %v", err)
	}
	if err = cc.Notify(&notify.Args{}, &notify.Reply{}); err != nil {
		t.Errorf("Error calling Notify: %v", err)
	}
	if err = cc.Write(&feglobal.WriteArgs{}, &feglobal.WriteReply{}); err != nil {
		t.Errorf("Error calling Write: %v", err)
	}
	if err = cc.PIR(&protocol.PIRArgs{}, &protocol.PIRReply{}); err != nil {
		t.Errorf("Error calling Read: %v", err)
	}

	c.Close()
	s.Close()
}
