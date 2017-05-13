package pir

import (
	"errors"
	"strings"

	"github.com/privacylab/talek/pir/pircpu"
)

// DB is a memory area for PIR computations shared with a PIR daemon.
type DB struct {
	DB    []byte
	shard *Shard
}

type pirReq struct {
	reqtype  int
	response chan []byte
}

// Server is a connection and state for a running PIR Server.
type Server struct {
	newshard   *func(int, []byte, string) Shard
	backing    string
	CellLength int
	CellCount  int
	BatchSize  int
	DB         *DB
}

// Connect opens a Server for communication with a unix-socket representating a
// running PIR Daemon.
func NewServer(backing string) (*Server, error) {
	server := new(Server)

	if strings.HasPrefix(backing, "cpu") {
		server.newshard = &pircpu.NewShard
	} else {
		return nil, errors.New("Backing " + backing + " is not known")
	}

	s.backing = backing
	return server, nil
}

// Disconnect closes a Server connection
func (s *Server) Disconnect() error {
	if s.DB != nil {
		s.DB.Free()
	}
}

// Configure sets the size of the DB and operational parameters.
func (s *Server) Configure(celllength int, cellcount int, batchsize int) error {
	s.BatchSize = batchsize
	s.CellCount = cellcount
	s.CellLength = celllength

	if s.CellCount%8 != 0 || s.CellLength%8 != 0 {
		return errors.New("invalid sizing of database; everything needs to be multiples of 8 bytes")
	}

	return nil
}

// GetDB provides direct access to the DB of the Server.
func (s *Server) GetDB() (*DB, error) {
	if s.CellCount == 0 || s.CellLength == 0 {
		return nil, errors.New("pir server unconfigured")
	}
	db := new(DB)

	db.DB = make([]byte, s.CellCount*s.CellLength)

	return db, nil
}

// SetDB updates the PIR Server Database
func (s *Server) SetDB(db *DB) error {
	if s.DB != nil {
		s.DB.shard.Free()
	}

	db.shard = s.newshard(s.CellLength, db.DB, s.backing)
	if db.shard == nil {
		return errors.New("Couldn't set DB")
	}
	s.DB = db
	return nil
}

// Free releases memory for a DB instance
func (db *DB) Free() error {
	if db.shard {
		db.shard.Free()
	}
	return nil
}

// Read makes a PIR request against the server.
func (s *Server) Read(masks []byte, responseChan chan []byte) error {
	if s.DB == nil || s.CellCount == 0 {
		return errors.New("db not configured")
	}

	if len(masks) != (s.CellCount*s.BatchSize)/8 {
		return errors.New("wrong mask length")
	}

	responses, err := s.DB.shard.Read(masks, len(masks))
	if err != nil {
		return err
	}
	responseChan <- responses

	return nil
}
