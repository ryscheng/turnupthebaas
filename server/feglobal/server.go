package feglobal

import (
	"sync"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/protocol/feglobal"
	"github.com/privacylab/talek/protocol/intvec"
	"github.com/privacylab/talek/protocol/notify"
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
	config     common.Config // Config

	// Thread-safe (organized by lock scope)
	lock       *sync.RWMutex
	snapshotID uint64
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
	s.config = config

	s.lock = &sync.RWMutex{}
	s.snapshotID = 0

	s.log.Info.Printf("feglobal.NewServer(%v) success\n", name)
	return s, nil
}

/**********************************
 * PUBLIC RPC METHODS (threadsafe)
 **********************************/

// GetInfo returns information about this server
func (s *Server) GetInfo(args *interface{}, reply *feglobal.GetInfoReply) error {
	tr := trace.New("FEGlobal", "GetInfo")
	defer tr.Finish()
	s.lock.RLock()

	reply.Err = ""
	reply.Name = s.name
	reply.SnapshotID = s.snapshotID

	s.lock.RUnlock()
	return nil
}

// GetIntVec returns the global interest vector
func (s *Server) GetIntVec(args *intvec.GetIntVecArgs, reply *intvec.GetIntVecReply) error {
	tr := trace.New("FEGlobal", "GetIntVec")
	defer tr.Finish()
	//s.lock.RLock()

	reply.Err = ""

	//s.lock.RUnlock()
	return nil

}

// Notify this server of a new snapshotID
func (s *Server) Notify(args *notify.Args, reply *notify.Reply) error {
	tr := trace.New("FEGlobal", "Notify")
	defer tr.Finish()
	//s.lock.RLock()

	//go s.NewLayout(args.Addr, args.SnapshotID)
	reply.Err = ""

	//s.lock.RUnlock()
	return nil
}

// Write stores a single message
func (s *Server) Write(args *feglobal.WriteArgs, reply *feglobal.WriteReply) error {
	tr := trace.New("FEGlobal", "Write")
	defer tr.Finish()

	// @todo
	reply.Err = ""

	return nil
}

// Read a batch of requests for a shard range
func (s *Server) Read(args *feglobal.ReadArgs, reply *feglobal.ReadReply) error {
	tr := trace.New("FEGlobal", "Read")
	defer tr.Finish()
	s.lock.RLock()

	// @todo
	reply.Err = ""

	s.lock.RUnlock()
	return nil
}

/**********************************
 * PUBLIC LOCAL METHODS (threadsafe)
 **********************************/

// Close shuts down the server
func (s *Server) Close() {
	//s.lock.Lock()

	if s.networkRPC != nil {
		s.networkRPC.Kill()
		s.networkRPC = nil
	}

	s.log.Info.Printf("%v.Close: success\n", s.name)
	//s.lock.Unlock()
}
