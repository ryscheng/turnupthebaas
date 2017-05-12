package coordinator

import (
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
	buildThreshold uint64
	buildInterval  time.Duration

	// Thread-safe
	config        atomic.Value  // Config
	commitLog     []*CommitArgs // Append and read only
	numNewCommits uint64
	timeLastBuild time.Time
	buildCount    uint64

	// Channels
	commitChan chan *CommitArgs
}

// NewServer creates a new Centralized talek server.
func NewServer(name string, config common.Config, buildThreshold uint64, buildInterval time.Duration) *Server {
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

	go s.processCommits()
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

// processCommits will read from s.commitChan and properly trigger work
func (s *Server) processCommits() {
	var commit *CommitArgs
	conf := s.config.Load().(common.Config)
	windowSize := conf.WindowSize()
	tick := time.After(s.buildInterval)

	triggerBuild := func() {
		// Garbage collect old items
		idx := 0
		if len(s.commitLog) > windowSize {
			idx = len(s.commitLog) - windowSize
		}
		s.commitLog = s.commitLog[idx:]
		// Spawn build goroutine
		go s.buildLayout(s.buildCount, s.commitLog[:])
		// Reset state
		s.numNewCommits = 0
		s.buildCount += 1
	}

	for {
		select {
		// Handle new commits
		case commit = <-s.commitChan:
			s.commitLog = append(s.commitLog, commit)
			s.numNewCommits += 1
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

func (s *Server) buildLayout(buildId uint64, config common.Config, commitLog []*CommitArgs) {
	tr := trace.New("Coordinator", "buildLayout")
	defer tr.Finish()

	// Construct cuckoo table layout
	data := make([]byte, config.NumBuckets*uint64(config.BucketDepth)*8)
	table := cuckoo.NewTable(string(buildId), config.NumBuckets, config.BucketDepth, config.DataSize, data, 0)

	// Construct global interest vector

	// Push layout to replicas

	// Push global interest vector to global frontends

}
