package replica

import (
	"sync"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/protocol/layout"
	"github.com/privacylab/talek/protocol/notify"
	"github.com/privacylab/talek/protocol/replica"
	"github.com/privacylab/talek/server"
	"golang.org/x/net/trace"
)

// Server is the main logic for replicas
type Server struct {
	/** Private State **/
	// Static
	log        *common.Logger
	name       string
	addr       string
	networkRPC *server.NetworkRPC

	// Thread-safe (organized by lock scope)
	lock         *sync.Mutex
	config       common.Config // Config
	snapshotID   uint64
	messages     map[uint64]*common.WriteArgs
	layoutAddr   string
	layoutClient *layout.Client

	// Channels
}

// NewServer creates a new replica server
func NewServer(name string, addr string, listenRPC bool, config common.Config) (*Server, error) {
	s := &Server{}
	s.log = common.NewLogger(name)
	s.name = name
	s.addr = addr
	s.networkRPC = nil
	if listenRPC {
		s.networkRPC = server.NewNetworkRPCAddr(s, addr)
	}

	s.lock = &sync.Mutex{}
	s.config = config
	s.snapshotID = 0
	s.messages = make(map[uint64]*common.WriteArgs)

	s.log.Info.Printf("replica.NewServer(%v) success\n", name)
	return s, nil
}

/**********************************
 * PUBLIC RPC METHODS (threadsafe)
 **********************************/

// GetInfo returns information about this server
func (s *Server) GetInfo(args *interface{}, reply *replica.GetInfoReply) error {
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
func (s *Server) Notify(args *notify.Args, reply *notify.Reply) error {
	tr := trace.New("Replica", "Notify")
	defer tr.Finish()
	s.lock.Lock()

	go s.GetLayout(args.Addr, args.SnapshotID)
	reply.Err = ""

	s.lock.Unlock()
	return nil
}

// Write stores a single message
func (s *Server) Write(args *common.WriteArgs, reply *common.WriteReply) error {
	tr := trace.New("Replica", "Write")
	defer tr.Finish()
	s.lock.Lock()

	s.messages[args.ID] = args
	reply.Err = ""

	s.lock.Unlock()
	return nil
}

// Read a batch of requests for a shard range
func (s *Server) Read(args *replica.ReadArgs, reply *replica.ReadReply) error {
	tr := trace.New("Replica", "Read")
	defer tr.Finish()
	s.lock.Lock()

	if s.snapshotID < args.SnapshotID {
		go s.GetLayout(s.layoutAddr, args.SnapshotID)
		reply.Err = "Need updated layout. Try again later."
		s.lock.Unlock()
		return nil
	}

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
	s.lock.Lock()

	if s.networkRPC != nil {
		s.networkRPC.Kill()
		s.networkRPC = nil
	}

	s.lock.Unlock()
}

// SetLayoutAddr will set the address and RPC client towards the server from which we get layouts
// Note: This will do nothing if addr is the same as we've seen before
func (s *Server) SetLayoutAddr(addr string) {
	s.lock.Lock()

	if s.layoutAddr != addr {
		if s.layoutClient != nil {
			s.layoutClient.Close()
			s.layoutClient = nil
		}
		s.layoutAddr = addr
		s.layoutClient = layout.NewClient(s.name, addr)
	}

	s.lock.Unlock()
}

// GetLayout will fetch the layout for a snapshotID and apply it locally
func (s *Server) GetLayout(addr string, snapshotID uint64) {
	tr := trace.New("Replica", "GetLayout")
	defer tr.Finish()

	// Try to establish an RPC client to server. Does nothing if addr is seen before
	s.SetLayoutAddr(addr)
	s.lock.Lock()

	var reply layout.GetLayoutReply
	args := &layout.GetLayoutArgs{
		SnapshotID: snapshotID,
		//ShardID: ,
		//NumShards: ,
	}
	// @todo + gc s.messages

	// Only set on success
	s.snapshotID = snapshotID

	s.lock.Unlock()
}

/**********************************
 * PRIVATE METHODS (single-threaded)
 **********************************/

/**********************************
 * HELPER FUNCTIONS
 **********************************/
