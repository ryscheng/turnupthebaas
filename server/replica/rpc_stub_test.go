package replica

import (
	"strconv"
	"testing"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/server"
)

const testPort = 9876

func TestRPCBasic(t *testing.T) {
	s, err := NewServer("test", testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	ss := server.NewNetworkRPC(s, testPort)
	c := NewRPCStub("test", "localhost:"+strconv.Itoa(testPort))
	var cc Interface
	cc = c

	if err = cc.GetInfo(nil, &GetInfoReply{}); err != nil {
		t.Errorf("Error calling GetInfo: %v", err)
	}
	if err = cc.GetCommonConfig(nil, &common.Config{}); err != nil {
		t.Errorf("Error calling GetCommonConfig: %v", err)
	}
	if err = cc.GetLayout(&GetLayoutArgs{}, &GetLayoutReply{}); err != nil {
		t.Errorf("Error calling GetLayout: %v", err)
	}
	if err = cc.GetIntVec(&GetIntVecArgs{}, &GetIntVecReply{}); err != nil {
		t.Errorf("Error calling GetIntVec: %v", err)
	}
	if err = cc.Commit(&CommitArgs{}, &CommitReply{}); err != nil {
		t.Errorf("Error calling Commit: %v", err)
	}

	c.Close()
	ss.Kill()
	s.Close()
}
