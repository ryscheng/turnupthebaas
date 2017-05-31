package replica

import (
	"testing"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/protocol/replica"
)

/********************************
 *** HELPER FUNCTIONS
 ********************************/

const testAddr = "localhost:9876"

func testConfig() common.Config {
	return common.Config{
		NumBuckets:         8,
		BucketDepth:        2,
		DataSize:           256,
		NumBucketsPerShard: 2,
		NumShardsPerGroup:  2,
		WriteInterval:      time.Minute,
		ReadInterval:       time.Minute,
		MaxLoadFactor:      0.50,
		BloomFalsePositive: 0.01,
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
