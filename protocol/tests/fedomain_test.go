package tests

import (
	"testing"

	protocol "github.com/privacylab/talek/protocol/fedomain"
	"github.com/privacylab/talek/protocol/feglobal"
	"github.com/privacylab/talek/protocol/layout"
	"github.com/privacylab/talek/protocol/notify"
	server "github.com/privacylab/talek/server/fedomain"
)

func TestFEDomain(t *testing.T) {
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
	if err = cc.GetLayout(&layout.Args{}, &layout.Reply{}); err != nil {
		t.Errorf("Error calling GetLayout: %v", err)
	}
	if err = cc.Notify(&notify.Args{}, &notify.Reply{}); err != nil {
		t.Errorf("Error calling Notify: %v", err)
	}
	if err = cc.Write(&feglobal.WriteArgs{}, &feglobal.WriteReply{}); err != nil {
		t.Errorf("Error calling Write: %v", err)
	}
	if err = cc.EncPIR(&feglobal.EncPIRArgs{}, &feglobal.ReadReply{}); err != nil {
		t.Errorf("Error calling EncPIR: %v", err)
	}

	c.Close()
	s.Close()
}
