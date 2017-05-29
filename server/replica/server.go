package replica

import (
	"sync"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/server/coordinator"
	"golang.org/x/net/trace"
)

// Server is the main logic for the central coordinator
type Server struct {
	/** Private State **/
	// Static
	log  *common.Logger
	name string

	// Thread-safe (organized by lock scope)
	lock       *sync.Mutex
	config     common.Config // Config
	snapshotID uint64



	// Channels
}

// NewServer creates a new replica server
func NewServer(name string, config common.Config) (*Server, error) {
	s := &Server{}
	s.log = common.NewLogger(name)
	s.name = name

	s.lock = &sync.Mutex{}
	s.config = config
	s.snapshotID = 0

	s.log.Info.Printf("replica.NewServer(%v) success\n", name)
	return s, nil
}

/**********************************
 * PUBLIC RPC METHODS (threadsafe)
 **********************************/

// GetInfo returns information about this server
func (s *Server) GetInfo(args *interface{}, reply *GetInfoReply) error {
	tr := trace.New("Replica", "GetInfo")
	defer tr.Finish()
	s.lock.Lock()

	reply.Err = ""
	reply.Name = s.name
	reply.SnapshotID = s.snapshotID

	s.lock.Unlock()
	return nil
}

// Notify this server of a new snapshotID
func (s *Server) Notify(args *coordinator.NotifyArgs, reply *coordinator.NotifyReply) error {
	tr := trace.New("Replica", "Notify")
	defer tr.Finish()
	s.lock.Lock()

  // @todo
	reply.Err = ""

	s.lock.Unlock()
	return nil
}

// Write stores a single message
func (s *Server) Write(args *common.WriteArgs, reply *common.WriteReply) error {
	tr := trace.New("Replica", "Write")
	defer tr.Finish()
	s.lock.Lock()

	reply.Err = ""

	s.lock.Unlock()
	return nil
}

// Read a batch of requests for a shard range
func (s *Server) Read(args *ReadArgs, reply *ReadReply) error {
	tr := trace.New("Replica", "Read")
	defer tr.Finish()
	s.lock.Lock()

  // @todo
	reply.Err = ""

	s.lock.Unlock()
	return nil
}

/**********************************
 * PUBLIC LOCAL METHODS (threadsafe)
 **********************************/

// Close shuts down the server
func (s *Server) Close() {
	s.log.Info.Printf("%v.Close: success", s.name)
}

/**********************************
 * PRIVATE METHODS (single-threaded)
 **********************************/

/**********************************
 * HELPER FUNCTIONS
 **********************************/
