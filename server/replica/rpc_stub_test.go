package replica

import (
	"strconv"
	"testing"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/server"
	"github.com/privacylab/talek/server/coordinator"
)

const testPort = 9876

func TestRPCBasic(t *testing.T) {
	s, err := NewServer("test", testConfig())
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
	if err = cc.Notify(&coordinator.NotifyArgs{}, &coordinator.NotifyReply{}); err != nil {
		t.Errorf("Error calling Notify: %v", err)
	}
	if err = cc.Write(&common.WriteArgs{}, &common.WriteReply{}); err != nil {
		t.Errorf("Error calling Write: %v", err)
	}
	if err = cc.Read(&ReadArgs{}, &ReadReply{}); err != nil {
		t.Errorf("Error calling Read: %v", err)
	}

	c.Close()
	ss.Kill()
	s.Close()
}
