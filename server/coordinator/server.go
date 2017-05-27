package coordinator

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/privacylab/bloom"
	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/cuckoo"
	"golang.org/x/net/trace"
)

// Server is the main logic for the central coordinator
type Server struct {
	/** Private State **/
	// Static
	log               *common.Logger
	name              string
	snapshotThreshold uint64
	snapshotInterval  time.Duration

	// Thread-safe (locked)
	lock          sync.Mutex
	config        common.Config // Config
	servers       []NotifyInterface
	commitLog     []*CommitArgs // Append and read only
	numNewCommits uint64
	snapshotCount uint64
	lastLayout    []uint64
	intVec        []uint64
	cuckooData    []byte
	cuckooTable   *cuckoo.Table

	// Channels
	notifyChan chan bool
}

// NewServer creates a new Centralized talek server.
func NewServer(name string, config common.Config, servers []NotifyInterface, snapshotThreshold uint64, snapshotInterval time.Duration) (*Server, error) {
	s := &Server{}
	s.log = common.NewLogger(name)
	s.name = name
	s.snapshotThreshold = snapshotThreshold
	s.snapshotInterval = snapshotInterval

	s.lock = sync.Mutex{}
	s.config = config
	if servers == nil {
		s.servers = make([]NotifyInterface, 0)
	} else {
		s.servers = servers
	}
	s.commitLog = make([]*CommitArgs, 0)
	s.numNewCommits = 0
	s.snapshotCount = 0
	s.lastLayout = nil
	s.intVec = nil
	s.cuckooData = make([]byte, config.NumBuckets*config.BucketDepth*uint64(common.IDSize))

	// Choose a random seed for the cuckoo table
	seedBytes := make([]byte, 8)
	_, err := rand.Read(seedBytes)
	if err != nil {
		s.log.Error.Printf("coordinator.NewServer(%v) error: %v", name, err)
		return nil, err
	}
	seed, _ := binary.Varint(seedBytes)
	s.cuckooTable = cuckoo.NewTable(name, config.NumBuckets, config.BucketDepth, config.DataSize, s.cuckooData, seed)
	s.notifyChan = make(chan bool)

	go s.loop()

	s.log.Info.Printf("coordinator.NewServer(%v) success\n", name)
	return s, nil
}

/**********************************
 * PUBLIC RPC METHODS (threadsafe)
 **********************************/

// GetInfo returns information about this server
func (s *Server) GetInfo(args *interface{}, reply *GetInfoReply) error {
	tr := trace.New("Coordinator", "GetInfo")
	defer tr.Finish()
	s.lock.Lock()

	reply.Err = ""
	reply.Name = s.name
	reply.SnapshotID = s.snapshotCount

	s.lock.Unlock()
	return nil
}

// GetCommonConfig returns the common global config
func (s *Server) GetCommonConfig(args *interface{}, reply *common.Config) error {
	tr := trace.New("Coordinator", "GetCommonConfig")
	defer tr.Finish()
	s.lock.Lock()

	*reply = s.config

	s.lock.Unlock()
	return nil
}

// GetLayout returns the layout for a shard
func (s *Server) GetLayout(args *GetLayoutArgs, reply *GetLayoutReply) error {
	tr := trace.New("Coordinator", "GetLayout")
	defer tr.Finish()
	s.lock.Lock()

	// Check for correct snapshot ID
	reply.SnapshotID = s.snapshotCount
	if args.SnapshotID != s.snapshotCount {
		reply.Err = "Invalid SnapshotID"
		s.lock.Unlock()
		return nil
	}

	shardSize := uint64(len(s.lastLayout)) / args.NumShards
	reply.Err = ""
	idx := args.ShardID * shardSize
	reply.Layout = s.lastLayout[idx:(idx + shardSize)]

	s.lock.Unlock()
	return nil
}

// GetIntVec returns the global interest vector
func (s *Server) GetIntVec(args *GetIntVecArgs, reply *GetIntVecReply) error {
	tr := trace.New("Coordinator", "GetIntVec")
	defer tr.Finish()
	s.lock.Lock()

	// Check for correct snapshot ID
	reply.SnapshotID = s.snapshotCount
	if args.SnapshotID != s.snapshotCount {
		reply.Err = "Invalid SnapshotID"
		s.lock.Unlock()
		return nil
	}

	reply.Err = ""
	reply.IntVec = s.intVec[:]

	s.lock.Unlock()
	return nil
}

// Commit accepts a single Write to commit. The
func (s *Server) Commit(args *CommitArgs, reply *CommitReply) error {
	tr := trace.New("Coordinator", "Commit")
	defer tr.Finish()
	reply.Err = ""

	s.lock.Lock()

	windowSize := s.config.WindowSize()
	// Garbage Collect old elements
	for uint64(len(s.commitLog)) >= windowSize {
		_ = s.cuckooTable.Remove(asCuckooItem(s.commitLog[0]))
		s.commitLog = s.commitLog[1:]
	}

	// Insert new item
	s.numNewCommits++
	s.commitLog = append(s.commitLog, args)
	ok, _ := s.cuckooTable.Insert(asCuckooItem(args))
	if !ok {
		s.log.Error.Fatalf("%v.processCommit failed to insert new element", s.name)
		return fmt.Errorf("Error inserting into cuckoo table")
	}

	s.lock.Unlock()

	// Do notifications in loop()
	s.notifyChan <- false
	return nil
}

/**
func (s *Server) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	tr := trace.New("Coordinator", "GetUpdates")
	defer tr.Finish()
}
**/

/**********************************
 * PUBLIC LOCAL METHODS (threadsafe)
 **********************************/

// Close shuts down the server
func (s *Server) Close() {
	s.log.Info.Printf("%v.Close: success", s.name)
}

// AddServer adds a server to the list that is notified on snapshot changes
func (s *Server) AddServer(server NotifyInterface) {
	s.lock.Lock()
	s.servers = append(s.servers, server)
	s.lock.Unlock()
}

// NotifySnapshot notifies the current cuckoo layout out
// If `force` is false, ignore when under a threshold
// Returns: true if snapshot was built, false if ignored
func (s *Server) NotifySnapshot(force bool) bool {
	s.lock.Lock()

	// Ignore if under threshold and not forcing
	if !force {
		if s.numNewCommits < s.snapshotThreshold {
			s.lock.Unlock()
			return false
		}
	}

	// Reset state
	s.numNewCommits = 0
	s.snapshotCount++

	// Construct global interest vector
	s.intVector = buildGlobalInterestVector(s.commitLog[:])

	// Copy the layout
	s.lastLayout = make([]uint64, len(s.cuckooData)/8)
	for i := 0; i < len(s.lastLayout); i++ {
		idx := i * 8
		s.lastLayout[i], _ = binary.Uvarint(s.cuckooData[idx:(idx + 8)])
	}

	// Sync with buildGlobalInterestVector goroutine
	// @todo

	// Send notifications
	go sendNotification(s.log, s.servers[:], s.snapshotCount)
	s.lock.Unlock()

	return true
}

/**********************************
 * PRIVATE METHODS (single-threaded)
 **********************************/
// Periodically call NotifySnapshot
func (s *Server) loop() {
	tick := time.After(s.snapshotInterval)
	for {
		select {
		// Called after Commit
		case <-s.notifyChan:
			ok := s.NotifySnapshot(false)
			if ok { // Re-up timer if snapshot built
				tick = time.After(s.snapshotInterval)
			}
			continue
		// Periodically trigger a snapshot
		case <-tick:
			s.NotifySnapshot(true)
			tick = time.After(s.snapshotInterval) // Re-up timer
			continue
		}
	}
}

/**********************************
 * HELPER FUNCTIONS
 **********************************/

// Converts a CommitArgs to a cuckoo.Item
func asCuckooItem(args *CommitArgs) *cuckoo.Item {
	itemData := make([]byte, common.IDSize)
	binary.PutUvarint(itemData, args.ID)
	return &cuckoo.Item{
		ID:      args.ID,
		Data:    itemData,
		Bucket1: args.Bucket1,
		Bucket2: args.Bucket2,
	}
}

// buildInterestVector traverses the commitLog and creates a global interest vector
// representing the elements
func buildInterestVector(commitLog []*CommitArgs) []uint64 {

	for _, c := range commitLog {

	}
}

func sendNotification(log *common.Logger, servers []NotifyInterface, snapshotID uint64) {
	args := &NotifyArgs{
		SnapshotID: snapshotID,
	}
	doNotify := func(s NotifyInterface, args *NotifyArgs) {
		err := s.Notify(args, &NotifyReply{})
		if err != nil {
			log.Error.Printf("sendNotification failed: %v", err)
		}
	}
	for _, s := range servers {
		go doNotify(s, args)
	}
}
