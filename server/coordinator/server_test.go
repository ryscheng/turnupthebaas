package coordinator

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/privacylab/talek/bloom"
	"github.com/privacylab/talek/common"
)

/********************************
 *** HELPER FUNCTIONS
 ********************************/

func testConfig() common.Config {
	return common.Config{
		NumBuckets:         1024,
		BucketDepth:        4,
		DataSize:           256,
		BloomFalsePositive: 0.01,
		WriteInterval:      time.Minute,
		ReadInterval:       time.Minute,
		MaxLoadFactor:      0.90,
	}
}

func afterEach(server *Server, mockChan []chan bool) {
	if server != nil {
		server.Close()
	}
	if mockChan != nil {
		for _, c := range mockChan {
			close(c)
		}
	}
}

type MockServer struct {
	Done chan bool
}

func NewMockServer() *MockServer {
	s := &MockServer{}
	s.Done = make(chan bool)
	return s
}

func (s *MockServer) Notify(args *NotifyArgs, reply *NotifyReply) error {
	s.Done <- true
	return fmt.Errorf("test")
}

func setupMocks(n int) ([]NotifyInterface, []chan bool) {
	servers := make([]NotifyInterface, n)
	channels := make([]chan bool, n)
	for i := 0; i < n; i++ {
		s := NewMockServer()
		servers[i] = s
		channels[i] = s.Done
	}
	return servers, channels
}

func newCommit() *CommitArgs {
	return &CommitArgs{
		ID:        rand.Uint64(),
		Bucket1:   rand.Uint64(),
		Bucket2:   rand.Uint64(),
		IntVecLoc: []uint64{rand.Uint64(), rand.Uint64()},
	}
}

/********************************
 *** TESTS
 ********************************/

func TestAsCuckooItem(t *testing.T) {
	numBuckets := uint64(10)
	args := newCommit()
	item := asCuckooItem(numBuckets, args)
	data, _ := binary.Uvarint(item.Data)
	if item.ID != args.ID ||
		item.Bucket1 != args.Bucket1%numBuckets ||
		item.Bucket2 != args.Bucket2%numBuckets ||
		len(item.Data) != common.IDSize ||
		data != args.ID {
		t.Errorf("Improper conversion: from CommitArgs=%v to cuckoo.Item=%v", args, item)
	}
}

func TestBuildInterestVector(t *testing.T) {
	commits := make([]*CommitArgs, 0)
	commits = append(commits, &CommitArgs{IntVecLoc: []uint64{15, 17}})
	commits = append(commits, &CommitArgs{IntVecLoc: []uint64{25, 27}})
	commits = append(commits, &CommitArgs{IntVecLoc: []uint64{35, 37, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65}})
	expectedLocs := []uint64{15, 17, 25, 27, 35, 37}
	intVec := buildInterestVector(1000, 0.01, commits)
	if !bloom.CheckLocations(intVec, expectedLocs) {
		t.Errorf("Interest Vector missing some bits")
	}
	// We never set this
	if bloom.CheckLocations(intVec, []uint64{0}) {
		t.Errorf("Interest Vector has extraneous bit")
	}
	// This should be outside of maximum number of bits allowed from the 3rd commit
	if bloom.CheckLocations(intVec, []uint64{65}) {
		t.Errorf("Interest Vector has extraneous bit")
	}
}

func TestSendNotification(t *testing.T) {
	numServers := 3
	mocks, channels := setupMocks(numServers)
	log := common.NewLogger("test")
	sendNotification(log, mocks, 10)
	for i := 0; i < numServers; i++ {
		select {
		case <-channels[i]:
			continue
		case <-time.After(time.Second):
			t.Errorf("Timed out before every server got a notification")
			break
		}
	}
	afterEach(nil, channels)
}

func TestNewServer(t *testing.T) {
	s, err := NewServer("test", testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	afterEach(s, nil)
}

func TestGetInfo(t *testing.T) {
	s, err := NewServer("test", testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	reply := &GetInfoReply{}
	err = s.GetInfo(nil, reply)
	if err != nil {
		t.Errorf("Error calling GetInfo: %v", err)
	}
	if reply.Err != "" || reply.Name != "test" || reply.SnapshotID != 0 {
		t.Errorf("Malformed reply from GetInfo: %v", reply)
	}
	afterEach(s, nil)
}

func TestGetCommonConfig(t *testing.T) {
	s, err := NewServer("test", testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	reply := &common.Config{}
	err = s.GetCommonConfig(nil, reply)
	if err != nil {
		t.Errorf("Error calling GetCommonConfig: %v", err)
	}
	if !reflect.DeepEqual(testConfig(), *reply) {
		t.Errorf("Malformed reply from GetCommonConfig: %v vs %v", reply, testConfig())
	}
	afterEach(s, nil)
}

func TestGetLayoutInvalidSnapshotID(t *testing.T) {
	s, err := NewServer("test", testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	args := &GetLayoutArgs{
		SnapshotID: 100,
		ShardID:    0,
		NumShards:  1,
	}
	reply := &GetLayoutReply{}
	if s.GetLayout(args, reply) != nil {
		t.Errorf("Error calling GetLayout: %v", err)
	}
	if reply.Err == "" {
		t.Errorf("GetLayout should have returned an error for invalid SnapshotID: %v", reply)
	}
	afterEach(s, nil)
}

func TestGetLayoutInvalidNumShards(t *testing.T) {
	s, err := NewServer("test", testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	args := &GetLayoutArgs{
		SnapshotID: 0,
		ShardID:    0,
		NumShards:  0,
	}
	reply := &GetLayoutReply{}
	if s.GetLayout(args, reply) != nil {
		t.Errorf("Error calling GetLayout: %v", err)
	}
	if reply.Err == "" {
		t.Errorf("GetLayout should have returned an error for invalid NumShards: %v", reply)
	}
	afterEach(s, nil)
}

func TestGetLayoutInvalidShardID(t *testing.T) {
	s, err := NewServer("test", testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	args := &GetLayoutArgs{
		SnapshotID: 0,
		ShardID:    4,
		NumShards:  4,
	}
	reply := &GetLayoutReply{}
	if s.GetLayout(args, reply) != nil {
		t.Errorf("Error calling GetLayout: %v", err)
	}
	if reply.Err == "" {
		t.Errorf("GetLayout should have returned an error for invalid ShardID: %v", reply)
	}
	afterEach(s, nil)
}

func TestGetLayoutEmpty(t *testing.T) {
	s, err := NewServer("test", testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	args := &GetLayoutArgs{
		SnapshotID: 0,
		ShardID:    0,
		NumShards:  1,
	}
	reply := &GetLayoutReply{}
	if s.GetLayout(args, reply) != nil {
		t.Errorf("Error calling GetLayout: %v", err)
	}
	if reply.Err != "" {
		t.Errorf("GetLayout should not return an error: %v", reply)
	}
	for _, v := range reply.Layout {
		if v != 0 {
			t.Errorf("GetLayout should return empty layout: %v", reply)
		}
	}
	afterEach(s, nil)
}

func TestGetIntVecInvalidSnapshotID(t *testing.T) {
	s, err := NewServer("test", testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	args := &GetIntVecArgs{SnapshotID: 100}
	reply := &GetIntVecReply{}
	if s.GetIntVec(args, reply) != nil {
		t.Errorf("Error calling GetIntVec: %v", err)
	}
	if reply.Err == "" {
		t.Errorf("GetIntVec should have returned an error for invalid SnapshotID: %v", reply)
	}
	afterEach(s, nil)
}

func TestGetIntVecEmpty(t *testing.T) {
	s, err := NewServer("test", testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	args := &GetIntVecArgs{SnapshotID: 0}
	reply := &GetIntVecReply{}
	if s.GetIntVec(args, reply) != nil {
		t.Errorf("Error calling GetIntVec: %v", err)
	}
	if reply.Err != "" || reply.SnapshotID != 0 {
		t.Errorf("GetIntVec should have succeeded: %v", reply)
	}
	for _, v := range reply.IntVec {
		if v != 0 {
			t.Errorf("GetIntVec should have returned an empty interest vector, %v", reply)
		}
	}
	afterEach(s, nil)
}

func TestCommit(t *testing.T) {
}

func TestSnapshot(t *testing.T) {
}

func TestSnapshotTimer(t *testing.T) {
}

func TestSnapshotThreshold(t *testing.T) {
}
