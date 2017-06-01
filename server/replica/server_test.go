package replica

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/protocol/layout"
	"github.com/privacylab/talek/protocol/replica"
	"github.com/privacylab/talek/server"
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

func randAddr() string {
	num := rand.Int()
	num %= 100
	num += 9800
	return "localhost:" + strconv.Itoa(num)
}

func testNewWrite(id uint64) *common.WriteArgs {
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

type MockServer struct {
	rpc    *server.NetworkRPC
	layout []uint64
	Done   chan *layout.GetLayoutArgs
}

func NewMockServer(addr string, l []uint64) *MockServer {
	s := &MockServer{}
	s.rpc = server.NewNetworkRPCAddr(s, addr)
	s.layout = l
	s.Done = make(chan *layout.GetLayoutArgs)
	return s
}

func (s *MockServer) Close() {
	s.rpc.Kill()
	close(s.Done)
}

func (s *MockServer) GetLayout(args *layout.GetLayoutArgs, reply *layout.GetLayoutReply) error {
	//s.Done <- args
	fmt.Println("!!!!!!")
	reply.Err = ""
	reply.SnapshotID = args.SnapshotID
	reply.Layout = s.layout
	return nil
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
	args := testNewWrite(10)
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

func TestGetSetLayoutAddr(t *testing.T) {
	s, err := NewServer("test", testAddr, false, testConfig(), 0, "cpu.0")
	if err != nil {
		t.Errorf("Error creating new server")
	}
	_, client0 := s.GetLayoutAddr()
	if client0 == nil {
		t.Errorf("GetLayout client should exist when server is created")
	}
	s.SetLayoutAddr(testAddr)
	addr1, client1 := s.GetLayoutAddr()
	if addr1 != testAddr || client1 == nil {
		t.Errorf("SetLayoutAddr didn't set the RPC client properly")
	}
	// Check that we don't create a new client when setting with same address
	s.SetLayoutAddr(testAddr)
	addr2, client2 := s.GetLayoutAddr()
	if addr1 != addr2 || client1 != client2 {
		t.Errorf("SetLayoutAddr should not have created a new client when setting with same address")
	}
	s.Close()
}

func TestGetLayout(t *testing.T) {
	mockAddr := randAddr()
	mock := NewMockServer(mockAddr, []uint64{0})
	s, err := NewServer("test", testAddr, false, testConfig(), 0, "cpu.0")
	if err != nil {
		t.Errorf("Error creating new server")
	}
	s.SetLayoutAddr(mockAddr)
	snapshotID, layout := s.GetLayout(1)
	if snapshotID != 1 || layout == nil {
		t.Errorf("GetLayout returns invalid results: snapshotID=%v, layout=%v", snapshotID, layout)
	}

	s.Close()
	mock.Close()
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
