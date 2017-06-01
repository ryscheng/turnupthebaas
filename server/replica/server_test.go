package replica

import (
	"encoding/binary"
	"math/rand"
	"testing"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/protocol/replica"
)

/********************************
 *** HELPER FUNCTIONS
 ********************************/

const testAddr = "localhost:9876"
const testNumBuckets = 8
const testDataSize = 256

func testConfig() common.Config {
	return common.Config{
		NumBuckets:         testNumBuckets,
		BucketDepth:        2,
		DataSize:           testDataSize,
		NumBucketsPerShard: 2,
		NumShardsPerGroup:  2,
		WriteInterval:      time.Minute,
		ReadInterval:       time.Minute,
		MaxLoadFactor:      0.50,
		BloomFalsePositive: 0.01,
	}
}

func testNewWrite() *common.WriteArgs {
	id := rand.Uint64()
	data := make([]byte, testDataSize)
	binary.PutUvarint(data, id)
	return &common.WriteArgs{
		ID:             id,
		Bucket1:        rand.Uint64() % testNumBuckets,
		Bucket2:        rand.Uint64() % testNumBuckets,
		Data:           data,
		InterestVector: []uint64{rand.Uint64(), rand.Uint64()},
	}
}

/********************************
 *** TESTS
 ********************************/

func TestNewServer(t *testing.T) {
	s, err := NewServer("test", testAddr, false, testConfig(), 0, "cpu.0")
	if err != nil {
		t.Errorf("Error creating new server")
	}
	s.Close()
}

func TestGetInfo(t *testing.T) {
	s, err := NewServer("test", testAddr, false, testConfig(), 0, "cpu.0")
	if err != nil {
		t.Errorf("Error creating new server")
	}
	reply := &replica.GetInfoReply{}
	err = s.GetInfo(nil, reply)
	if err != nil {
		t.Errorf("Error calling GetInfo: %v", err)
	}
	if reply.Err != "" || reply.Name != "test" || reply.SnapshotID != 0 {
		t.Errorf("Malformed reply from GetInfo: %v", reply)
	}
	s.Close()
}

func TestNotify(t *testing.T) {
	// @todo
}

func TestWrite(t *testing.T) {
	s, err := NewServer("test", testAddr, false, testConfig(), 0, "cpu.0")
	if err != nil {
		t.Errorf("Error creating new server")
	}
	reply := &common.WriteReply{}
	args := testNewWrite()
	err = s.Write(args, reply)
	if err != nil {
		t.Errorf("Error calling Write: %v", err)
	}
	if reply.Err != "" {
		t.Errorf("Malformed reply from Write: %v", reply)
	}
	s.Close()
}

func TestRead(t *testing.T) {
	// @todo
}

func TestSetLayoutAddr(t *testing.T) {
	s, err := NewServer("test", testAddr, false, testConfig(), 0, "cpu.0")
	if err != nil {
		t.Errorf("Error creating new server")
	}
	addr, client := s.GetLayoutAddr()
	if client == nil {
		t.Errorf("GetLayout client should exist when server is created")
	}
	s.SetLayoutAddr(testAddr)
	addr, client = s.GetLayoutAddr()
	if addr != testAddr || client == nil {
		t.Errorf("SetLayoutAddr didn't set the RPC client properly")
	}
	// Check that we don't create a new client when setting with same address
	s.SetLayoutAddr(testAddr)
	addr2, client2 := s.GetLayoutAddr()
	if addr != addr2 || client != client2 {
		t.Errorf("SetLayoutAddr should not have created a new client when setting with same address")
	}
	s.Close()
}

func TestGetLayout(t *testing.T) {
	// @todo
}

func TestApplyLayout(t *testing.T) {
	// @todo
}

func TestSetShards(t *testing.T) {
	// @todo
}

func TestNewLayout(t *testing.T) {
	// @todo
}
