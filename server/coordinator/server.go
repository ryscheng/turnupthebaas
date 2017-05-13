package coordinator

import (
	"encoding/binary"
	"sync/atomic"
	"time"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/cuckoo"
	"golang.org/x/net/trace"
)

// Server is the main logic for the central coordinator
type Server struct {
	/** Private State **/
	// Static
	log            *common.Logger
	name           string
	buildThreshold int64
	buildInterval  time.Duration

	// Thread-safe
	config        atomic.Value  // Config
	commitLog     []*CommitArgs // Append and read only
	numNewCommits int64
	timeLastBuild time.Time
	buildCount    int64

	// Channels
	commitChan chan *CommitArgs
}

// NewServer creates a new Centralized talek server.
func NewServer(name string, config common.Config, buildThreshold int64, buildInterval time.Duration) *Server {
	s := &Server{}
	s.log = common.NewLogger(name)
	s.name = name
	s.buildThreshold = buildThreshold
	s.buildInterval = buildInterval

	s.config.Store(config)
	s.commitLog = make([]*CommitArgs, 0)
	s.numNewCommits = 0
	s.timeLastBuild = time.Now()
	s.buildCount = 0

	go s.loop()
	return s
}

/**********************************
 * PUBLIC RPC METHODS (threadsafe)
 **********************************/

// GetInfo returns information about this server
func (s *Server) GetInfo(args *interface{}, reply *GetInfoReply) error {
	tr := trace.New("Coordinator", "GetInfo")
	defer tr.Finish()
	reply.Err = ""
	reply.Name = s.name
	return nil
}

// GetCommonConfig returns the common global config
func (s *Server) GetCommonConfig(args *interface{}, reply *common.Config) error {
	tr := trace.New("Coordinator", "GetCommonConfig")
	defer tr.Finish()
	config := s.config.Load().(common.Config)
	*reply = config
	return nil
}

// Commit accepts a single Write to commit. The
func (s *Server) Commit(args *CommitArgs, reply *CommitReply) error {
	tr := trace.New("Coordinator", "Commit")
	defer tr.Finish()
	s.commitChan <- args
	reply.Err = ""
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
	close(s.commitChan)
}

/**********************************
 * PRIVATE METHODS (single-threaded)
 **********************************/

func (s *Server) loop() {
	var commit *CommitArgs
	conf := s.config.Load().(common.Config)
	windowSize := conf.WindowSize()
	tick := time.After(s.buildInterval)

	triggerBuild := func() {
		// Garbage collect old items
		logLength := uint64(len(s.commitLog))
		idx := uint64(0)
		if logLength > windowSize {
			idx = logLength - windowSize
		}
		s.commitLog = s.commitLog[idx:]
		// Spawn build goroutine
		go s.buildLayout(s.buildCount, conf, s.commitLog[:])
		go s.buildInterestVec(s.buildCount, conf, s.commitLog[:])
		// Reset state
		s.numNewCommits = 0
		s.buildCount++
	}

	for {
		select {
		// Handle new commits
		case commit = <-s.commitChan:
			s.commitLog = append(s.commitLog, commit)
			s.numNewCommits++
			// Trigger build if over threshold
			if s.numNewCommits > s.buildThreshold {
				triggerBuild()
			}
			continue
		// Periodically trigger a build
		case <-tick:
			triggerBuild()
			tick = time.After(s.buildInterval) // Re-up timer
			continue
		}
	}
}

func (s *Server) buildLayout(buildID int64, config common.Config, commitLog []*CommitArgs) {
	tr := trace.New("Coordinator", "buildLayout")
	defer tr.Finish()

	// Construct cuckoo table layout
	var item *cuckoo.Item
	var ok bool
	itemData := make([]byte, common.IDSize)
	data := make([]byte, config.NumBuckets*config.BucketDepth*uint64(common.IDSize))
	table := cuckoo.NewTable(string(buildID), config.NumBuckets, config.BucketDepth, config.DataSize, data, 0)
	for _, elt := range commitLog {
		binary.PutUvarint(itemData, elt.ID)
		item = &cuckoo.Item{
			ID:      elt.ID,
			Data:    itemData,
			Bucket1: elt.Bucket1,
			Bucket2: elt.Bucket2,
		}
		ok, _ = table.Insert(item)
		if !ok {
			s.log.Error.Fatalf("%v.buildLayout failed to construct cuckoo layout", s.name)
		}
	}
	s.log.Info.Printf("%v\n", data)

	// Push layout to replicas
	// @todo

}

func (s *Server) buildInterestVec(buildID int64, config common.Config, commitLog []*CommitArgs) {
	// Construct global interest vector
	// @todo

	// Push global interest vector to global frontends
	// @todo

}
