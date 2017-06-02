package fedomain

import (
	"sync"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/protocol/fedomain"
	"github.com/privacylab/talek/protocol/feglobal"
	"github.com/privacylab/talek/protocol/layout"
	"github.com/privacylab/talek/protocol/notify"
	"github.com/privacylab/talek/server"
	"golang.org/x/net/trace"
)

// Server is the main logic for trust domain frontends
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

// NewServer creates a new trust domain frontend server
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

	s.log.Info.Printf("fedomain.NewServer(%v) success\n", name)
	return s, nil
}

/**********************************
 * PUBLIC RPC METHODS (threadsafe)
 **********************************/

// GetInfo returns information about this server
func (s *Server) GetInfo(args *interface{}, reply *fedomain.GetInfoReply) error {
	tr := trace.New("FEDomain", "GetInfo")
	defer tr.Finish()
	s.lock.RLock()

	reply.Err = ""
	reply.Name = s.name
	reply.SnapshotID = s.snapshotID

	s.lock.RUnlock()
	return nil
}

// GetLayout returns the layout
func (s *Server) GetLayout(args *layout.GetLayoutArgs, reply *layout.GetLayoutReply) error {
	tr := trace.New("FEDomain", "GetLayout")
	defer tr.Finish()
	//s.lock.RLock()

	reply.Err = ""

	//s.lock.RUnlock()
	return nil

}

// Notify this server of a new snapshotID
func (s *Server) Notify(args *notify.Args, reply *notify.Reply) error {
	tr := trace.New("FEDomain", "Notify")
	defer tr.Finish()
	//s.lock.RLock()

	//go s.NewLayout(args.Addr, args.SnapshotID)
	reply.Err = ""

	//s.lock.RUnlock()
	return nil
}

// Write stores a single message
func (s *Server) Write(args *feglobal.WriteArgs, reply *feglobal.WriteReply) error {
	tr := trace.New("FEDomain", "Write")
	defer tr.Finish()

	// @todo
	reply.Err = ""

	return nil
}

// EncPIR processes a single encrypted PIR requests
func (s *Server) EncPIR(args *feglobal.EncPIRArgs, reply *feglobal.ReadReply) error {
	tr := trace.New("FEDomain", "EncPIR")
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
