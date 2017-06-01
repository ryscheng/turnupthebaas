package replica

import (
	"encoding/binary"
	"math/rand"
	"reflect"
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
	rpc        *server.NetworkRPC
	err        string
	snapshotID uint64
	layout     []uint64
	Done       chan *layout.GetLayoutArgs
}

func NewMockServer(addr string, err string, snapshotID uint64, l []uint64) *MockServer {
	s := &MockServer{}
	s.rpc = server.NewNetworkRPCAddr(s, addr)
	s.err = err
	s.snapshotID = snapshotID
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
	// Make sure GetLayout eventually terminates
	if args.SnapshotID == s.snapshotID && s.err == layout.ErrorInvalidSnapshotID {
		reply.Err = ""
	} else {
		reply.Err = s.err
	}

	reply.SnapshotID = s.snapshotID
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
	// @todo notify, check GetLayout called, check shards set
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

func TestWriteBadData(t *testing.T) {
	s, err := NewServer("test", testAddr, false, testConfig(), 0, "cpu.0")
	if err != nil {
		t.Errorf("Error creating new server")
	}
	reply := &common.WriteReply{}
	args := testNewWrite(10)
	args.Data = make([]byte, 1) // Invalid DataSize!
	err = s.Write(args, reply)
	if err != nil {
		t.Errorf("Error calling Write: %v", err)
	}
	if reply.Err == "" {
		t.Errorf("Write should have returned InvalidDataSize: %v", reply)
	}
	s.Close()
}

func TestPIR(t *testing.T) {
	// @todo - setshard, do read, check correctness
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
	mockSnapshotID := uint64(1)
	mockLayout := []uint64{0, 1, 2}
	mock := NewMockServer(mockAddr, "", mockSnapshotID, mockLayout)
	s, err := NewServer("test", testAddr, false, testConfig(), 0, "cpu.0")
	if err != nil {
		t.Errorf("Error creating new server")
	}
	s.SetLayoutAddr(mockAddr)
	snapshotID, layout := s.GetLayout(mockSnapshotID)
	if snapshotID != mockSnapshotID || !reflect.DeepEqual(layout, mockLayout) {
		t.Errorf("GetLayout returns invalid results: snapshotID=%v, layout=%v", snapshotID, layout)
	}
	s.Close()
	mock.Close()
}

func TestGetLayoutRPCFail(t *testing.T) {
	mockAddr := randAddr()
	mockSnapshotID := uint64(1)
	s, err := NewServer("test", testAddr, false, testConfig(), 0, "cpu.0")
	if err != nil {
		t.Errorf("Error creating new server")
	}
	s.SetLayoutAddr(mockAddr)
	snapshotID, layout := s.GetLayout(mockSnapshotID)
	if snapshotID != mockSnapshotID || layout != nil {
		t.Errorf("GetLayout should have failed with no server to contact")
	}
	s.Close()
}

func TestGetLayoutErrorInvalidSnapshotID(t *testing.T) {
	mockAddr := randAddr()
	mockSnapshotID := uint64(1)
	mockLayout := []uint64{0, 1, 2}
	mock := NewMockServer(mockAddr, layout.ErrorInvalidSnapshotID, mockSnapshotID, mockLayout)
	s, err := NewServer("test", testAddr, false, testConfig(), 0, "cpu.0")
	if err != nil {
		t.Errorf("Error creating new server")
	}
	s.SetLayoutAddr(mockAddr)
	snapshotID, layout := s.GetLayout(0)
	if snapshotID != mockSnapshotID || !reflect.DeepEqual(layout, mockLayout) {
		t.Errorf("GetLayout should have eventually succeeded when starting with invalid snapshotID")
	}
	s.Close()
	mock.Close()
}

func TestGetLayoutErrorInvalidIndex(t *testing.T) {
	mockAddr := randAddr()
	mockSnapshotID := uint64(1)
	mockLayout := []uint64{0, 1, 2}
	mock := NewMockServer(mockAddr, layout.ErrorInvalidIndex, mockSnapshotID, mockLayout)
	s, err := NewServer("test", testAddr, false, testConfig(), 0, "cpu.0")
	if err != nil {
		t.Errorf("Error creating new server")
	}
	s.SetLayoutAddr(mockAddr)
	snapshotID, layout := s.GetLayout(mockSnapshotID)
	if snapshotID != mockSnapshotID || layout != nil {
		t.Errorf("GetLayout should have failed with ErrorInvalidIndex")
	}
	s.Close()
	mock.Close()
}

func TestApplyLayout(t *testing.T) {
	// @todo - Do writes, apply layout, check correctness
}

func TestApplyLayoutMissingMessage(t *testing.T) {
	// @todo -
}

func TestSetShards(t *testing.T) {
	// @todo - Construct a shard, set
}

func TestNewLayout(t *testing.T) {
	// @todo - writes, Notify, PIR
}
