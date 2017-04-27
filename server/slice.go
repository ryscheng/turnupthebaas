package server

import (
	"fmt"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/pir"
)

// One item schedulable for placement in this slice
type entry struct {
	Data            []byte
	Bucket1         int
	Bucket2         int
	CurrentLocation int
}

// SliceOperation represents a serialized request to read from, or change the
// slice of the database.
type SliceOperation struct {
	seqNo       uint64
	updateState map[int][]uint64
	read        []common.PirArgs
}

// responses to reads are sent back to the requesting thread
type response struct {
}

// Slice represents a slice of a talek replica's Database. It take in parallel
// streams of read and write updates, and handles translating those into queries
// on an underlying PIR daemon.
type Slice struct {
	log *common.Logger
	*pir.Server
	*pir.DB
	dead int

	// The first bucket managed by this slice
	offset uint64
	// The number of buckets managed by this slice
	buckets uint64

	config *Config

	entries map[uint64]entry

	ops       chan *SliceOperation
	responses chan []byte
}

// NewSlice creates a new slice to handle reads and writes
func NewSlice(socket string, offset uint64, buckets uint64, config *Config, responses chan []byte) *Slice {
	s := &Slice{}
	s.log = common.NewLogger(fmt.Sprintf("slice[%d-%d]", offset, buckets))
	s.config = config
	s.offset = offset
	s.buckets = buckets
	s.ops = make(chan *SliceOperation, 10)
	s.responses = responses

	pirServer, err := pir.Connect(socket)
	if err != nil {
		s.log.Error.Fatalf("Could not connect to PIR: %v", err)
		return nil
	}
	s.Server = pirServer
	if err = s.Server.Configure(config.DataSize*config.BucketDepth, int(buckets), config.ReadBatch); err != nil {
		s.log.Error.Fatalf("Could not configure PIR: %v", err)
		return nil
	}

	db, err := pirServer.GetDB()
	if err != nil {
		s.log.Error.Fatalf("Could not allocate DB: %v", err)
		return nil
	}
	s.DB = db

	// Set initial DB
	s.Server.SetDB(s.DB)

	go s.process()
	return s
}

// Handle an operation sent from the controller.
func (s *Slice) Handle(op *SliceOperation) {
	s.ops <- op
}

// Writes have less-strict serialization requirements than reads &
// database updates, and are sent separately. An 'Apply' operation will
// block until the sequence range is fully available.
func (s *Slice) Write(seqNo uint64, item entry) {
	s.entries[seqNo] = item
}

// Close shuts down this component
func (s *Slice) Close() {
	s.log.Info.Printf("Graceful shutdown started.")
	if s.dead > 0 {
		s.log.Warn.Printf("Slice already shutting down.")
		return
	}
	s.dead = 1
}

func (s *Slice) process() {
	defer s.DB.Free()
	defer s.Server.Disconnect()

	var op *SliceOperation

	for s.dead == 0 {
		select {
		case op = <-s.ops:
			if op.read != nil {
				// TODO: turn req vectors in a concatinated string.
				s.Server.Read(op.read[0].RequestVector, s.responses)
			} else {
				// TODO: handle database update
			}
		}
	}
}
