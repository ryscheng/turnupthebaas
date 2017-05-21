package pir

import (
	"errors"

	"github.com/privacylab/talek/pir/pirinterface"
)

// DB is a memory area for PIR computations shared with a PIR daemon.
type DB struct {
	DB    []byte
	shard pirinterface.Shard
}

type pirReq struct {
	reqtype  int
	response chan []byte
}

// Server is a connection and state for a running PIR Server.
type Server struct {
	newshard   func(int, []byte, string) pirinterface.Shard
	backing    string
	CellLength int
	CellCount  int
	BatchSize  int
	DB         *DB
}

// NewServer creates a Server for communication
func NewServer(backing string) (*Server, error) {
	server := new(Server)
	server.backing = backing

	if cons := pirinterface.GetBacking(backing); cons != nil {
		server.newshard = cons
		return server, nil
	}

	return nil, errors.New("Backing " + backing + " is not known")
}

// Disconnect closes a Server connection
func (s *Server) Disconnect() error {
	if s.DB != nil {
		s.DB.Free()
	}
	return nil
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
		s.DB.Free()
	}

	shardMemory := make([]byte, len(db.DB))
	copy(shardMemory[:], db.DB[:])
	db.shard = s.newshard(s.CellLength, shardMemory, s.backing)
	if db.shard == nil {
		return errors.New("Couldn't set DB")
	}
	s.DB = db
	return nil
}

// Free releases memory for a DB instance
func (db *DB) Free() error {
	if db.shard != nil {
		db.shard.Free()
		db.shard = nil
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

	responses, err := s.DB.shard.Read(masks, s.CellCount/8)
	if err != nil {
		return err
	}
	responseChan <- responses

	return nil
}
