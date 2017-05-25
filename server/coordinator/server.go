package coordinator

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/cuckoo"
	"golang.org/x/net/trace"
)

// Server is the main logic for the central coordinator
type Server struct {
	/** Private State **/
	// Static
	log           *common.Logger
	name          string
	pushThreshold uint64
	pushInterval  time.Duration

	// Thread-safe
	lock          sync.Mutex
	config        common.Config // Config
	commitLog     []*CommitArgs // Append and read only
	numNewCommits uint64
	pushCount     uint64
	lastLayout    []uint64
	cuckooData    []byte
	cuckooTable   *cuckoo.Table

	// Channels
	// - public
	//LayoutChan chan []byte
	//InterestVecChan chan
	// - private
}

// NewServer creates a new Centralized talek server.
func NewServer(name string, config common.Config, pushThreshold uint64, pushInterval time.Duration) (*Server, error) {
	s := &Server{}
	s.log = common.NewLogger(name)
	s.name = name
	s.pushThreshold = pushThreshold
	s.pushInterval = pushInterval

	s.lock = sync.Mutex{}
	s.config = config
	s.commitLog = make([]*CommitArgs, 0)
	s.numNewCommits = 0
	s.pushCount = 0
	s.lastLayout = nil
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
	//s.log.Info.Printf("%v\n", data)

	s.lock.Unlock()

	s.PushLayout(false)
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

// PushLayout pushes the current cuckoo layout out
// If `force` is false, ignore when under a threshold
func (s *Server) PushLayout(force bool) {
	s.lock.Lock()

	// Ignore if under threshold and not forcing
	if !force {
		if s.numNewCommits < s.pushThreshold {
			return
		}
	}

	// Reset state
	s.numNewCommits = 0
	s.pushCount++
	// Copy the layout
	s.lastLayout = make([]uint64, len(s.cuckooData)/8)
	for i := 0; i < len(s.lastLayout); i++ {
		idx := i * 8
		s.lastLayout[i], _ = binary.Uvarint(s.cuckooData[idx:(idx + 8)])
	}
	go sendLayout(s.pushCount, s.lastLayout[:], s.commitLog[:])
	s.lock.Unlock()
}

/**********************************
 * PRIVATE METHODS (single-threaded)
 **********************************/
// Periodically call PushLayout
func (s *Server) loop() {
	tick := time.After(s.pushInterval)
	for {
		select {
		// Periodically trigger a config push
		case <-tick:
			s.PushLayout(true)
			tick = time.After(s.pushInterval) // Re-up timer
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

func buildGlobalInterestVector() {
	// @todo
}

func sendLayout(pushID uint64, layout []uint64, commitLog []*CommitArgs) {
	// @todo
	// Push layout to replicas

	// Construct global interest vector
	// intVec := buildGlobalInterestVector(commitLog)
	// Push global interest vector to global frontends
	// @todo

}
