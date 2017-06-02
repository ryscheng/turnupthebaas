package tests

import (
	"testing"
	"time"

	"github.com/privacylab/talek/common"
	protocol "github.com/privacylab/talek/protocol/coordinator"
	"github.com/privacylab/talek/protocol/intvec"
	"github.com/privacylab/talek/protocol/layout"
	server "github.com/privacylab/talek/server/coordinator"
)

func TestCoordinator(t *testing.T) {
	testAddr := randAddr()
	s, err := server.NewServer("test", testAddr, true, testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	c := protocol.NewClient("test", testAddr)
	var cc protocol.Interface
	cc = c

	if err = cc.GetInfo(nil, &protocol.GetInfoReply{}); err != nil {
		t.Errorf("Error calling GetInfo: %v", err)
	}
	if err = cc.GetCommonConfig(nil, &common.Config{}); err != nil {
		t.Errorf("Error calling GetCommonConfig: %v", err)
	}
	if err = cc.GetLayout(&layout.Args{}, &layout.Reply{}); err != nil {
		t.Errorf("Error calling GetLayout: %v", err)
	}
	if err = cc.GetIntVec(&intvec.Args{}, &intvec.Reply{}); err != nil {
		t.Errorf("Error calling GetIntVec: %v", err)
	}
	if err = cc.Commit(&protocol.CommitArgs{}, &protocol.CommitReply{}); err != nil {
		t.Errorf("Error calling Commit: %v", err)
	}

	c.Close()
	s.Close()
}
