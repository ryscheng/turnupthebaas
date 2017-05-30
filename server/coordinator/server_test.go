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
	"github.com/privacylab/talek/protocol/coordinator"
	"github.com/privacylab/talek/protocol/intvec"
	"github.com/privacylab/talek/protocol/layout"
	"github.com/privacylab/talek/protocol/notify"
)

const testAddr = "localhost:9876"

/********************************
 *** HELPER FUNCTIONS
 ********************************/

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

func afterEach(server *Server, mockChan []chan *notify.Args) {
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
	Done chan *notify.Args
}

func NewMockServer() *MockServer {
	s := &MockServer{}
	s.Done = make(chan *notify.Args)
	return s
}

func (s *MockServer) Notify(args *notify.Args, reply *notify.Reply) error {
	s.Done <- args
	return fmt.Errorf("test")
}

func setupMocks(n int) ([]notify.Interface, []chan *notify.Args) {
	servers := make([]notify.Interface, n)
	channels := make([]chan *notify.Args, n)
	for i := 0; i < n; i++ {
		s := NewMockServer()
		servers[i] = s
		channels[i] = s.Done
	}
	return servers, channels
}

func newCommit() *coordinator.CommitArgs {
	return &coordinator.CommitArgs{
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
		len(item.Data) != coordinator.IDSize ||
		data != args.ID {
		t.Errorf("Improper conversion: from CommitArgs=%v to cuckoo.Item=%v", args, item)
	}
}

func TestBuildInterestVector(t *testing.T) {
	commits := make([]*coordinator.CommitArgs, 0)
	commits = append(commits, &coordinator.CommitArgs{IntVecLoc: []uint64{15, 17}})
	commits = append(commits, &coordinator.CommitArgs{IntVecLoc: []uint64{25, 27}})
	commits = append(commits, &coordinator.CommitArgs{IntVecLoc: []uint64{35, 37, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65}})
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
	sendNotification(log, mocks, 10, "")
	// Wait for all notifications

	for i := 0; i < numServers; i++ {
		select {
		case args := <-channels[i]:
			if args.SnapshotID != 10 {
				t.Errorf("Wrong SnapshotID in notify")
			}
			continue
		case <-time.After(time.Second):
			t.Errorf("Timed out before every server got a notification")
			break
		}
	}
	afterEach(nil, channels)
}

func TestNewServer(t *testing.T) {
	s, err := NewServer("test", testAddr, false, testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	afterEach(s, nil)
}

func TestGetInfo(t *testing.T) {
	s, err := NewServer("test", testAddr, false, testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	reply := &coordinator.GetInfoReply{}
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
	s, err := NewServer("test", testAddr, false, testConfig(), nil, 5, time.Hour)
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
	s, err := NewServer("test", testAddr, false, testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	args := &layout.GetLayoutArgs{
		SnapshotID: 100,
		ShardID:    0,
		NumShards:  1,
	}
	reply := &layout.GetLayoutReply{}
	if s.GetLayout(args, reply) != nil {
		t.Errorf("Error calling GetLayout: %v", err)
	}
	if reply.Err == "" {
		t.Errorf("GetLayout should have returned an error for invalid SnapshotID: %v", reply)
	}
	afterEach(s, nil)
}

func TestGetLayoutInvalidNumShards(t *testing.T) {
	s, err := NewServer("test", testAddr, false, testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	args := &layout.GetLayoutArgs{
		SnapshotID: 0,
		ShardID:    0,
		NumShards:  0,
	}
	reply := &layout.GetLayoutReply{}
	if s.GetLayout(args, reply) != nil {
		t.Errorf("Error calling GetLayout: %v", err)
	}
	if reply.Err == "" {
		t.Errorf("GetLayout should have returned an error for invalid NumShards: %v", reply)
	}
	afterEach(s, nil)
}

func TestGetLayoutInvalidShardID(t *testing.T) {
	s, err := NewServer("test", testAddr, false, testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	args := &layout.GetLayoutArgs{
		SnapshotID: 0,
		ShardID:    4,
		NumShards:  4,
	}
	reply := &layout.GetLayoutReply{}
	if s.GetLayout(args, reply) != nil {
		t.Errorf("Error calling GetLayout: %v", err)
	}
	if reply.Err == "" {
		t.Errorf("GetLayout should have returned an error for invalid ShardID: %v", reply)
	}
	afterEach(s, nil)
}

func TestGetLayoutEmpty(t *testing.T) {
	s, err := NewServer("test", testAddr, false, testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	args := &layout.GetLayoutArgs{
		SnapshotID: 0,
		ShardID:    0,
		NumShards:  1,
	}
	reply := &layout.GetLayoutReply{}
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
	s, err := NewServer("test", testAddr, false, testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	args := &intvec.GetIntVecArgs{SnapshotID: 100}
	reply := &intvec.GetIntVecReply{}
	if s.GetIntVec(args, reply) != nil {
		t.Errorf("Error calling GetIntVec: %v", err)
	}
	if reply.Err == "" {
		t.Errorf("GetIntVec should have returned an error for invalid SnapshotID: %v", reply)
	}
	afterEach(s, nil)
}

func TestGetIntVecEmpty(t *testing.T) {
	s, err := NewServer("test", testAddr, false, testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	args := &intvec.GetIntVecArgs{SnapshotID: 0}
	reply := &intvec.GetIntVecReply{}
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
	s, err := NewServer("test", testAddr, false, testConfig(), nil, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	args := newCommit()
	reply := &coordinator.CommitReply{}
	if s.Commit(args, reply) != nil {
		t.Errorf("Error calling Commit: %v", err)
	}
	if reply.Err != "" {
		t.Errorf("Commit should have succeeded: %v", reply)
	}
	afterEach(s, nil)
}

func TestAddServer(t *testing.T) {
	numServers := 3
	mocks, channels := setupMocks(numServers)
	s, err := NewServer("test", testAddr, false, testConfig(), mocks, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	for _, m := range mocks {
		s.AddServer(m)
	}
	afterEach(s, channels)
}

func TestSnapshot(t *testing.T) {
	numServers := 3
	config := testConfig()
	mocks, channels := setupMocks(numServers)
	s, err := NewServer("test", testAddr, false, config, mocks, 5, time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	// Add a commit
	commit := newCommit()
	if s.Commit(commit, &coordinator.CommitReply{}) != nil {
		t.Errorf("Error calling Commit: %v", err)
	}
	// Force a snapshot
	s.NotifySnapshot(true)
	// Wait for all notifications
	for i := 0; i < numServers; i++ {
		select {
		case args := <-channels[i]:
			if args.SnapshotID != 1 {
				t.Errorf("Wrong SnapshotID in notify")
			}
			continue
		case <-time.After(time.Second):
			t.Errorf("Timed out before every server got a notification")
			break
		}
	}
	// Check layout
	layoutArgs := &coordinator.GetLayoutArgs{SnapshotID: 1, ShardID: 0, NumShards: 1}
	layoutReply := &coordinator.GetLayoutReply{}
	if s.GetLayout(layoutArgs, layoutReply) != nil {
		t.Errorf("Error calling GetLayout: %v", err)
	}
	if layoutReply.Err != "" || layoutReply.SnapshotID != 1 {
		t.Errorf("GetLayout error: %v", layoutReply)
	}
	for i, id := range layoutReply.Layout {
		if id == commit.ID {
			bucket := uint64(i) / config.BucketDepth
			if bucket != commit.Bucket1 && bucket != commit.Bucket2 {
				t.Errorf("Invalid layout. Commit not in correct location")
			}
		}
	}
	// Check interest vector
	intVecArgs := &coordinator.GetIntVecArgs{SnapshotID: 1}
	intVecReply := &coordinator.GetIntVecReply{}
	if s.GetIntVec(intVecArgs, intVecReply) != nil {
		t.Errorf("Error calling GetIntVec: %v", err)
	}
	if intVecReply.Err != "" || intVecReply.SnapshotID != 1 {
		t.Errorf("GetIntVec error: %v", intVecReply)
	}
	numBits, _ := bloom.EstimateParameters(config.WindowSize(), config.BloomFalsePositive)
	intVec := bloom.From(numBits, intVecReply.IntVec)
	if !bloom.Equal(intVec, bloom.SetLocations(bloom.NewBitSet(numBits), commit.IntVecLoc)) {
		t.Errorf("Invalid interest vector. Commit not included")
	}

	afterEach(s, channels)
}

func TestSnapshotThreshold(t *testing.T) {
	numServers := 3
	snapshotThreshold := 32
	mocks, channels := setupMocks(numServers)
	s, err := NewServer("test", testAddr, false, testConfig(), mocks, uint64(snapshotThreshold), time.Hour)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	// Add commits
	for i := 0; i < snapshotThreshold; i++ {
		if s.Commit(newCommit(), &coordinator.CommitReply{}) != nil {
			t.Errorf("Error calling Commit: %v", err)
		}
	}
	// Wait for all notifications
	for i := 0; i < numServers; i++ {
		select {
		case args := <-channels[i]:
			if args.SnapshotID != 1 {
				t.Errorf("Wrong SnapshotID in notify")
			}
			continue
		case <-time.After(time.Second):
			t.Errorf("Timed out before every server got a notification")
			break
		}
	}
	afterEach(s, channels)
}

func TestSnapshotTimer(t *testing.T) {
	numServers := 3
	mocks, channels := setupMocks(numServers)
	s, err := NewServer("test", testAddr, false, testConfig(), mocks, 5, 10*time.Millisecond)
	if err != nil {
		t.Errorf("Error creating new server")
	}
	if s.Commit(newCommit(), &coordinator.CommitReply{}) != nil {
		t.Errorf("Error calling Commit: %v", err)
	}
	// Wait for all notifications
	for i := 0; i < numServers; i++ {
		select {
		case args := <-channels[i]:
			if args.SnapshotID != 1 {
				t.Errorf("Wrong SnapshotID in notify")
			}
			continue
		case <-time.After(time.Second):
			t.Errorf("Timed out before every server got a notification")
			break
		}
	}
	afterEach(s, channels)

}
