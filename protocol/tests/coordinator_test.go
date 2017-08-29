package tests

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/privacylab/talek/common"
	protocol "github.com/privacylab/talek/protocol/coordinator"
	server "github.com/privacylab/talek/server/coordinator"
)

const testAddr = "localhost:9876"

func testConfig() common.Config {
	return common.Config{
		NumBuckets:         8,
		BucketDepth:        2,
		DataSize:           256,
		BloomFalsePositive: 0.01,
		WriteInterval:      time.Minute,
		ReadInterval:       time.Minute,
		MaxLoadFactor:      0.50,
	}
}

func TestRPCBasic(t *testing.T) {
	s, err := server.NewServer("test", testAddr, testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	testAddrTCP, _ := net.ResolveTCPAddr("tcp4", testAddr)
	l, err := net.ListenTCP("tcp4", testAddrTCP)
	if err != nil {
		t.Errorf("Could not bind to test address")
	}
	go http.Serve(l, s)
	c := protocol.NewClient("test", testAddr)
	var cc protocol.Interface
	cc = c

	if err = cc.GetInfo(nil, &protocol.GetInfoReply{}); err != nil {
		t.Errorf("Error calling GetInfo: %v", err)
	}
	if err = cc.GetCommonConfig(nil, &common.Config{}); err != nil {
		t.Errorf("Error calling GetCommonConfig: %v", err)
	}
	if err = cc.GetLayout(&protocol.GetLayoutArgs{}, &protocol.GetLayoutReply{}); err != nil {
		t.Errorf("Error calling GetLayout: %v", err)
	}
	if err = cc.GetIntVec(&protocol.GetIntVecArgs{}, &protocol.GetIntVecReply{}); err != nil {
		t.Errorf("Error calling GetIntVec: %v", err)
	}
	if err = cc.Commit(&protocol.CommitArgs{}, &protocol.CommitReply{}); err != nil {
		t.Errorf("Error calling Commit: %v", err)
	}

	c.Close()
	s.Close()
}
