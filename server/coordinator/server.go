package coordinator

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/privacylab/talek/bloom"
	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/cuckoo"
	"github.com/privacylab/talek/protocol/coordinator"
	"github.com/privacylab/talek/protocol/intvec"
	"github.com/privacylab/talek/protocol/layout"
	"github.com/privacylab/talek/protocol/notify"
	"github.com/privacylab/talek/server"
	"golang.org/x/net/trace"
)

// Server is the main logic for the central coordinator
type Server struct {
	/** Private State **/
	// Static
	log               *common.Logger
	name              string
	addr              string
	networkRPC        *server.NetworkRPC
	snapshotThreshold uint64
	snapshotInterval  time.Duration

	// Thread-safe (locked)
	lock          *sync.RWMutex
	config        common.Config // Config
	servers       []notify.Interface
	commitLog     []*coordinator.CommitArgs // Append and read only
	numNewCommits uint64
	snapshotCount uint64
	lastLayout    []uint64
	intVec        []uint64
	cuckooData    []byte
	cuckooTable   *cuckoo.Table

	// Channels
	notifyChan chan bool
	closeChan  chan bool
}

// NewServer creates a new Talek centralized coordinator server
func NewServer(name string, addr string, listenRPC bool, config common.Config, servers []notify.Interface, snapshotThreshold uint64, snapshotInterval time.Duration) (*Server, error) {
	s := &Server{}
	s.log = common.NewLogger(name)
	s.name = name
	s.addr = addr
	s.networkRPC = nil
	if listenRPC {
		s.networkRPC = server.NewNetworkRPCAddr(s, addr)
	}
	s.snapshotThreshold = snapshotThreshold
	s.snapshotInterval = snapshotInterval

	s.lock = &sync.RWMutex{}
	s.config = config
	if servers == nil {
		s.servers = make([]notify.Interface, 0)
	} else {
		s.servers = servers
	}
	s.commitLog = make([]*coordinator.CommitArgs, 0)
	s.numNewCommits = 0
	s.snapshotCount = 0
	s.lastLayout = make([]uint64, config.NumBuckets*config.BucketDepth)
	s.intVec = buildInterestVector(config.WindowSize(), config.BloomFalsePositive, s.commitLog[:]).Bytes()
	s.cuckooData = make([]byte, config.NumBuckets*config.BucketDepth*uint64(coordinator.IDSize))

	// Choose a random seed for the cuckoo table
	seedBytes := make([]byte, 8)
	_, err := rand.Read(seedBytes)
	if err != nil {
		s.log.Error.Printf("coordinator.NewServer(%v) error: %v", name, err)
		return nil, err
	}
	seed, _ := binary.Varint(seedBytes)
	s.cuckooTable = cuckoo.NewTable(name, config.NumBuckets, config.BucketDepth, uint64(coordinator.IDSize), s.cuckooData, seed)
	// Should not be possible
	if s.cuckooTable == nil {
		err := fmt.Errorf("Invalid cuckoo table parameters")
		s.log.Error.Printf("coordinator.NewServer(%v) error: %v", name, err)
		return nil, err
	}
	s.notifyChan = make(chan bool)
	s.closeChan = make(chan bool)

	go s.loop()

	s.log.Info.Printf("coordinator.NewServer(%v) success\n", name)
	return s, nil
}

/**********************************
 * PUBLIC RPC METHODS (threadsafe)
 **********************************/

// GetInfo returns information about this server
func (s *Server) GetInfo(args *interface{}, reply *coordinator.GetInfoReply) error {
	tr := trace.New("Coordinator", "GetInfo")
	defer tr.Finish()
	s.lock.RLock()

	reply.Err = ""
	reply.Name = s.name
	reply.SnapshotID = s.snapshotCount

	s.lock.RUnlock()
	return nil
}

// GetCommonConfig returns the common global config
func (s *Server) GetCommonConfig(args *interface{}, reply *common.Config) error {
	tr := trace.New("Coordinator", "GetCommonConfig")
	defer tr.Finish()
	s.lock.RLock()

	*reply = s.config

	s.lock.RUnlock()
	return nil
}

// GetLayout returns the layout for a shard
func (s *Server) GetLayout(args *layout.Args, reply *layout.Reply) error {
	tr := trace.New("Coordinator", "GetLayout")
	defer tr.Finish()
	s.lock.RLock()

	// Check for correct snapshot ID
	reply.SnapshotID = s.snapshotCount
	if args.SnapshotID != s.snapshotCount {
		reply.Err = layout.ErrorInvalidSnapshotID
		s.lock.RUnlock()
		return nil
	}

	// Check non-zero NumSplit
	if args.NumSplit < 1 {
		reply.Err = layout.ErrorInvalidNumSplit
		s.lock.RUnlock()
		return nil
	}

	shardSize := uint64(len(s.lastLayout)) / args.NumSplit
	idx := args.Index * shardSize
	if idx < 0 || (idx+shardSize) > uint64(len(s.lastLayout)) {
		reply.Err = layout.ErrorInvalidIndex
	} else {
		reply.Err = ""
		reply.Layout = s.lastLayout[idx:(idx + shardSize)]
	}

	s.lock.RUnlock()
	return nil
}

// GetIntVec returns the global interest vector
func (s *Server) GetIntVec(args *intvec.Args, reply *intvec.Reply) error {
	tr := trace.New("Coordinator", "GetIntVec")
	defer tr.Finish()
	s.lock.RLock()

	// Check for correct snapshot ID
	reply.SnapshotID = s.snapshotCount
	if args.SnapshotID != s.snapshotCount {
		reply.Err = "Invalid SnapshotID"
		s.lock.RUnlock()
		return nil
	}

	reply.Err = ""
	reply.IntVec = s.intVec[:]

	s.lock.RUnlock()
	return nil
}

// Commit accepts a single Write to commit without data. Used to maintain the cuckoo table
func (s *Server) Commit(args *coordinator.CommitArgs, reply *coordinator.CommitReply) error {
	tr := trace.New("Coordinator", "Commit")
	defer tr.Finish()
	reply.Err = ""

	s.lock.Lock()

	windowSize := s.config.WindowSize()
	// Garbage Collect old elements
	for uint64(len(s.commitLog)) >= windowSize {
		_ = s.cuckooTable.Remove(asCuckooItem(s.config.NumBuckets, s.commitLog[0]))
		s.commitLog = s.commitLog[1:]
	}

	// Insert new item
	s.numNewCommits++
	s.commitLog = append(s.commitLog, args)
	ok, _ := s.cuckooTable.Insert(asCuckooItem(s.config.NumBuckets, args))
	if !ok {
		s.log.Error.Fatalf("%v.processCommit failed to insert new element", s.name)
		return fmt.Errorf("Error inserting into cuckoo table")
	}

	s.lock.Unlock()

	// Do notifications in loop()
	s.notifyChan <- false
	return nil
}

/**********************************
 * PUBLIC LOCAL METHODS (threadsafe)
 **********************************/

// Close shuts down the server
func (s *Server) Close() {
	s.log.Info.Printf("%v.Close: success", s.name)
	//s.lock.Lock()

	s.closeChan <- true
	if s.networkRPC != nil {
		s.networkRPC.Kill()
		s.networkRPC = nil
	}

	//s.lock.Unlock()
}

// AddServer adds a server to the list that is notified on snapshot changes
func (s *Server) AddServer(server notify.Interface) {
	s.lock.Lock()
	s.servers = append(s.servers, server)
	s.log.Info.Printf("%v.AddServer() success\n", s.name)
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
	s.intVec = buildInterestVector(s.config.WindowSize(), s.config.BloomFalsePositive, s.commitLog[:]).Bytes()

	// Copy the layout
	for i := 0; i < len(s.lastLayout); i++ {
		idx := i * 8
		s.lastLayout[i], _ = binary.Uvarint(s.cuckooData[idx:(idx + 8)])
	}

	// Sync with buildGlobalInterestVector goroutine
	// @todo when this happens in parallel

	// Send notifications only if there are servers
	if s.servers != nil && len(s.servers) > 0 {
		go sendNotification(s.log, s.servers[:], s.snapshotCount, s.addr)
	}

	s.log.Info.Printf("%v.NotifySnapshot() success\n", s.name)
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
		// Stop the loop
		case <-s.closeChan:
			return
		}
	}
}

/**********************************
 * HELPER FUNCTIONS
 **********************************/

// Converts a CommitArgs to a cuckoo.Item
func asCuckooItem(numBuckets uint64, args *coordinator.CommitArgs) *cuckoo.Item {
	itemData := make([]byte, coordinator.IDSize)
	binary.PutUvarint(itemData, args.ID)
	return &cuckoo.Item{
		ID:      args.ID,
		Data:    itemData,
		Bucket1: args.Bucket1 % numBuckets,
		Bucket2: args.Bucket2 % numBuckets,
	}
}

// buildInterestVector traverses the commitLog and creates a global interest vector
// representing the elements.
// Bloom filter stores n elements with fp false positive rate
func buildInterestVector(n uint64, fp float64, commitLog []*coordinator.CommitArgs) *bloom.BitSet {
	bits, numHash := bloom.EstimateParameters(n, fp)
	intVec := bloom.NewBitSet(bits)
	for _, c := range commitLog {
		// Truncate interest vectors to first numHash bits
		if uint64(len(c.IntVecLoc)) > numHash {
			intVec = bloom.SetLocations(intVec, c.IntVecLoc[:numHash])
		} else {
			intVec = bloom.SetLocations(intVec, c.IntVecLoc[:])
		}
	}
	return intVec
}

func sendNotification(log *common.Logger, servers []notify.Interface, snapshotID uint64, addr string) {
	args := &notify.Args{
		SnapshotID: snapshotID,
		Addr:       addr,
	}
	doNotify := func(l *common.Logger, s notify.Interface, args *notify.Args) {
		err := s.Notify(args, &notify.Reply{})
		if err != nil {
			l.Error.Printf("sendNotification failed: %v", err)
		}
	}
	for _, s := range servers {
		go doNotify(log, s, args)
	}
}
