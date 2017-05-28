package coordinator

import (
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/privacylab/talek/bloom"
	"github.com/privacylab/talek/common"
)

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
	close(s.Done)
	return fmt.Errorf("test")
}

func setupServers(n int) ([]NotifyInterface, []chan bool) {
	servers := make([]NotifyInterface, n)
	channels := make([]chan bool, n)
	for i := 0; i < n; i++ {
		s := NewMockServer()
		servers[i] = s
		channels[i] = s.Done
	}
	return servers, channels
}

func TestAsCuckooItem(t *testing.T) {
	args := &CommitArgs{
		ID:        32,
		Bucket1:   5,
		Bucket2:   7,
		IntVecLoc: []uint64{9, 11},
	}
	item := asCuckooItem(args)
	data, _ := binary.Uvarint(item.Data)
	if item.ID != args.ID ||
		item.Bucket1 != args.Bucket1 ||
		item.Bucket2 != args.Bucket2 ||
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
	servers, channels := setupServers(numServers)
	log := common.NewLogger("test")
	sendNotification(log, servers, 10)
	for i := 0; i < numServers; i++ {
		select {
		case <-channels[i]:
			continue
		case <-time.After(time.Second):
			t.Errorf("Timed out before every server got a notification")
			break
		}
	}
}
