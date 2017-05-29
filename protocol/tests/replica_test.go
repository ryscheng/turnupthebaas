package tests

import (
	"testing"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/protocol/notify"
	protocol "github.com/privacylab/talek/protocol/replica"
	server "github.com/privacylab/talek/server/replica"
)

func TestReplica(t *testing.T) {
	testAddr := randAddr()
	s, err := server.NewServer("test", testAddr, true, testConfig())
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
	if err = cc.Write(&common.WriteArgs{}, &common.WriteReply{}); err != nil {
		t.Errorf("Error calling Write: %v", err)
	}
	if err = cc.Read(&protocol.ReadArgs{}, &protocol.ReadReply{}); err != nil {
		t.Errorf("Error calling Read: %v", err)
	}

	c.Close()
	s.Close()
}
