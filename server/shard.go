package server

import (
	"fmt"
	"math/rand"
	"os"
	"sync/atomic"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/cuckoo"
	"github.com/privacylab/talek/pir"
)

func getSocket() string {
	if os.Getenv("PIR_SOCKET") != "" {
		fmt.Printf("Testing against running pird at %s.\n", os.Getenv("PIR_SOCKET"))
		return os.Getenv("PIR_SOCKET")
	}
	return fmt.Sprintf("pirtest%d.socket", rand.Int())
}

// Shard represents a single shard of the PIR database.
// It runs a thread handling processing of incoming requests. It is
// responsible for the logic of placing writes into memory and managing a cuckoo
// hash table for doing so, and of driving the PIR daemon.
type Shard struct {
	// Private State
	log  *common.Logger
	name string

	*pir.Server
	*pir.DB
	dead int

	Entries []cuckoo.Item
	*cuckoo.Table

	config atomic.Value // Config

	// Channels
	writeChan        chan *common.WriteArgs
	readChan         chan *DecodedBatchReadRequest
	outstandingReads chan chan *common.BatchReadReply
	readReplies      chan []byte
	syncChan         chan int

	sinceFlip        int
	outstandingLimit int
}

// DecodedBatchReadRequest represents a set of PIR args from clients.
// The Centralized server manages decoding of read requests to the client and
// applying the PadSeed for the TrustDomain
type DecodedBatchReadRequest struct {
	Args      []common.PirArgs
	ReplyChan chan *common.BatchReadReply
}

// NewShard creates an interface to a PIR daemon at socket, using a given
// server configuration for sizing and locating data.
func NewShard(name string, socket string, config Config) *Shard {
	s := &Shard{}
	s.log = common.NewLogger(name)
	s.name = name

	s.config.Store(config)
	s.writeChan = make(chan *common.WriteArgs)
	s.readChan = make(chan *DecodedBatchReadRequest)
	s.syncChan = make(chan int)
	s.outstandingReads = make(chan chan *common.BatchReadReply, 5)
	s.readReplies = make(chan []byte)

	// TODO: per-server config of where the local PIR socket is.
	pirServer, err := pir.Connect(socket)
	if err != nil {
		s.log.Error.Fatalf("Could not connect to pir back end: %v", err)
		return nil
	}
	s.Server = pirServer
	err = s.Server.Configure(config.Config.DataSize*config.Config.BucketDepth, int(config.Config.NumBuckets), config.ReadBatch)
	if err != nil {
		s.log.Error.Fatalf("Could not start PIR back end with correct parameters: %v", err)
		return nil
	}

	db, err := pirServer.GetDB()
	if err != nil {
		s.log.Error.Fatalf("Could not allocate DB region: %v", err)
		return nil
	}
	s.DB = db
	//Set initial DB
	s.Server.SetDB(s.DB)

	// TODO: rand seed
	s.Table = cuckoo.NewTable(name+"-Table", int(config.Config.NumBuckets), config.Config.BucketDepth, config.Config.DataSize, db.DB, 0)
	s.Entries = make([]cuckoo.Item, 0, int(config.Config.NumBuckets)*config.Config.BucketDepth)

	//TODO: should be a parameter in globalconfig
	s.outstandingLimit = int(float32(config.Config.NumBuckets*uint64(config.Config.BucketDepth)) * 0.50)

	go s.processReads()
	go s.processReplies()
	go s.processWrites()
	return s
}

/** PUBLIC METHODS (threadsafe) **/

func (s *Shard) Write(args *common.WriteArgs) error {
	s.log.Trace.Println("Write: ")
	s.writeChan <- args
	return nil
}

// GetUpdates returns the current global interest vector of changed items.
func (s *Shard) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	s.log.Trace.Println("GetUpdates: ")
	// @TODO
	reply.Err = ""
	//reply.InterestVector =
	return nil
}

// BatchRead performs a read of a set of client requests against the database.
func (s *Shard) BatchRead(args *DecodedBatchReadRequest) {
	s.readChan <- args
}

// Close shuts down the database.
func (s *Shard) Close() {
	s.log.Info.Printf("Graceful shutdown of shard.")
	if s.dead > 0 {
		s.log.Warn.Printf("Shard already stopped / stopping.")
		return
	}
	s.dead = 1
	s.writeChan <- nil
	s.readChan <- nil
	<-s.syncChan
}

/** PRIVATE METHODS (singlethreaded) **/
func (s *Shard) processReads() {
	// The read thread searializs all access to the underlying DB
	var batchReadReq *DecodedBatchReadRequest

	defer s.DB.Free()
	defer s.Server.Disconnect()
	conf := s.config.Load().(Config)
	for {
		select {
		case batchReadReq = <-s.readChan:
			if batchReadReq == nil {
				s.log.Info.Printf("Read loop closed.")
				s.syncChan <- 0
				return
			}
			s.batchRead(batchReadReq, conf)
			continue
		case <-s.syncChan:
			s.Server.SetDB(s.DB)
		}
	}
}

func (s *Shard) processReplies() {
	var outputChannel chan *common.BatchReadReply
	conf := s.config.Load().(Config)
	itemLength := conf.DataSize * conf.BucketDepth

	for {
		select {
		case reply := <-s.readReplies:
			// get the corresponding read request.
			outputChannel = <-s.outstandingReads

			response := &common.BatchReadReply{Err: "", Replies: make([]common.ReadReply, conf.ReadBatch)}
			for i := 0; i < conf.ReadBatch; i++ {
				response.Replies[i].Data = reply[i*itemLength : (i+1)*itemLength]
				//TODO: reply.GlobalSeqNo
			}
			outputChannel <- response
		}
	}
}

func (s *Shard) processWrites() {
	var writeReq *common.WriteArgs
	conf := s.config.Load().(Config)
	for {
		select {
		case writeReq = <-s.writeChan:
			if writeReq == nil {
				return
			}

			itm := asCuckooItem(writeReq)
			s.Entries = append(s.Entries, *itm)
			ok, evicted := s.Table.Insert(itm)
			// No longer need this pointer.
			itm.Data = nil
			if !ok || len(s.Entries) > int(float32(int(conf.Config.NumBuckets)*conf.Config.BucketDepth)*conf.Config.MaxLoadFactor) {
				s.evictOldItems()
			}
			if evicted != nil {
				ok, evicted = s.Table.Insert(evicted)
				if !ok || evicted != nil {
					s.log.Error.Fatalf("Consistency violation: lost an in-window DB item.")
				}
			}
			s.sinceFlip++

			// Trigger to swap to next DB.
			if s.sinceFlip > s.outstandingLimit {
				s.ApplyWrites()
			}
		}
	}
}

// ApplyWrites will enque a command to apply any outstanding writes to the
// database to be seen by subsequent reads.
// TODO: should take a sequence number to do a better job of consistent
// interleaving with enqueued reads.
func (s *Shard) ApplyWrites() {
	s.syncChan <- 1
	s.sinceFlip = 0
}

func (s *Shard) evictOldItems() {
	conf := s.config.Load().(Config)
	toRemove := int(float32(int(conf.Config.NumBuckets)*conf.Config.BucketDepth) * conf.Config.LoadFactorStep)
	if toRemove >= len(s.Entries) {
		toRemove = len(s.Entries) - 1
	}
	for i := 0; i < toRemove; i++ {
		s.Table.Remove(&s.Entries[i])
	}
	s.Entries = s.Entries[toRemove:]
}

func asCuckooItem(wa *common.WriteArgs) *cuckoo.Item {
	//TODO: cuckoo should continue int64 sized buckets if needed.
	return &cuckoo.Item{ID: int(wa.GlobalSeqNo), Data: wa.Data, Bucket1: int(wa.Bucket1), Bucket2: int(wa.Bucket2)}
}

func (s *Shard) batchRead(req *DecodedBatchReadRequest, conf Config) {
	s.log.Trace.Printf("batchRead: enter\n")

	// Run PIR
	reqlength := int(conf.Config.NumBuckets) / 8
	pirvector := make([]byte, reqlength*conf.ReadBatch)

	if len(req.Args) != conf.ReadBatch {
		s.log.Info.Printf("Read operation failed: incorrect number of reads.")
		req.ReplyChan <- &common.BatchReadReply{Err: fmt.Sprintf("Invalid batch size.")}
		return
	}

	for i := 0; i < conf.ReadBatch; i++ {
		reqVector := req.Args[i].RequestVector
		copy(pirvector[reqlength*i:reqlength*(i+1)], reqVector)
	}
	err := s.Server.Read(pirvector, s.readReplies)
	if err != nil {
		s.log.Error.Fatalf("Reading from PIR Server failed: %v", err)
		req.ReplyChan <- &common.BatchReadReply{Err: fmt.Sprintf("Failed to read: %v", err)}
		return
	}
	s.outstandingReads <- req.ReplyChan

	s.log.Trace.Printf("batchRead: exit\n")
}
