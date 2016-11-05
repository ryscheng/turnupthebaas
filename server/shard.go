package server

import (
	"fmt"
	"github.com/ryscheng/pdb/common"
	"github.com/ryscheng/pdb/cuckoo"
	"github.com/ryscheng/pdb/pir"
	"math/rand"
	"os"
	"sync/atomic"
)

func getSocket() string {
	if os.Getenv("PIR_SOCKET") != "" {
		fmt.Printf("Testing against running pird at %s.\n", os.Getenv("PIR_SOCKET"))
		return os.Getenv("PIR_SOCKET")
	}
	return fmt.Sprintf("pirtest%d.socket", rand.Int())
}

/**
 * Handles a shard of the data
 * Goroutines:
 * - 1x processRequests()
 */
type Shard struct {
	// Private State
	log  *common.Logger
	name string

	*pir.PirServer
	*pir.PirDB

	Entries []cuckoo.Item
	*cuckoo.Table

	globalConfig atomic.Value //common.GlobalConfig

	// Channels
	writeChan chan *common.WriteArgs
	readChan  chan *common.BatchReadRequest
	syncChan  chan int

	sinceFlip        int
	outstandingLimit int
}

func NewShard(name string, socket string, globalConfig common.GlobalConfig) *Shard {
	s := &Shard{}
	s.log = common.NewLogger(name)
	s.name = name

	s.globalConfig.Store(globalConfig)
	s.writeChan = make(chan *common.WriteArgs)
	s.readChan = make(chan *common.BatchReadRequest)
	s.syncChan = make(chan int)

	// TODO: per-server config of where the local PIR socket is.
	pirServer, err := pir.Connect(socket)
	if err != nil {
		s.log.Error.Fatalf("Could not connect to pir back end: %v", err)
		return nil
	}
	s.PirServer = pirServer
	err = s.PirServer.Configure(globalConfig.DataSize*globalConfig.BucketDepth, int(globalConfig.NumBuckets), globalConfig.ReadBatch)
	if err != nil {
		s.log.Error.Fatalf("Could not start PIR back end with correct parameters: %v", err)
		return nil
	}

	db, err := pirServer.GetDB()
	if err != nil {
		s.log.Error.Fatalf("Could not allocate DB region: %v", err)
		return nil
	}
	s.PirDB = db
	//Set initial DB
	s.PirServer.SetDB(s.PirDB)

	// TODO: rand seed
	s.Table = cuckoo.NewTable(name+"-Table", int(globalConfig.NumBuckets), globalConfig.BucketDepth, globalConfig.DataSize, db.DB, 0)
	s.Entries = make([]cuckoo.Item, 0, int(globalConfig.NumBuckets)*globalConfig.BucketDepth)

	//TODO: should be a parameter in globalconfig
	s.outstandingLimit = int(float32(globalConfig.NumBuckets) * 0.10)

	go s.processReads()
	go s.processWrites()
	return s
}

/** PUBLIC METHODS (threadsafe) **/
func (s *Shard) Ping(args *common.PingArgs, reply *common.PingReply) error {
	s.log.Info.Println("Ping: " + args.Msg + ", ... Pong")
	reply.Err = ""
	reply.Msg = "PONG"
	return nil
}

func (s *Shard) Write(args *common.WriteArgs, reply *common.WriteReply) error {
	s.log.Trace.Println("Write: ")
	s.writeChan <- args
	reply.Err = ""
	return nil
}

func (s *Shard) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	s.log.Trace.Println("GetUpdates: ")
	// @TODO
	reply.Err = ""
	//reply.InterestVector =
	return nil
}

func (s *Shard) BatchRead(args *common.BatchReadArgs, replyChan chan *common.BatchReadReply) error {
	s.log.Trace.Println("BatchRead: ")
	batchReq := &common.BatchReadRequest{args, replyChan}
	s.readChan <- batchReq
	return nil
}

func (s *Shard) Close() {
	s.log.Info.Printf("Graceful shutdown of shard.")
	s.writeChan <- nil
	s.readChan <- nil
	<-s.syncChan
	s.log.Info.Printf("Caller thread knows read loop closed.")
}

/** PRIVATE METHODS (singlethreaded) **/
func (s *Shard) processReads() {
	// The read thread searializs all access to the underlying DB
	var batchReadReq *common.BatchReadRequest

	defer s.PirDB.Free()
	defer s.PirServer.Disconnect()
	for {
		select {
		case batchReadReq = <-s.readChan:
			if batchReadReq == nil {
				s.log.Info.Printf("Read loop closed.")
				s.syncChan <- 0
				return
			}
			s.batchRead(batchReadReq)
			continue
		case <-s.syncChan:
			s.PirServer.SetDB(s.PirDB)
		}
	}
}

func (s *Shard) processWrites() {
	var writeReq *common.WriteArgs
	conf := s.globalConfig.Load().(common.GlobalConfig)
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
			if !ok || len(s.Entries) > int(float32(int(conf.NumBuckets)*conf.BucketDepth)*conf.MaxLoadFactor) {
				s.evictOldItems()
			}
			if evicted != nil {
				ok, evicted = s.Table.Insert(evicted)
				if !ok || evicted != nil {
					s.log.Error.Fatalf("Consistency violation: lost an in-window DB item.")
				}
			}
			s.sinceFlip += 1

			// Trigger to swap to next DB.
			// TODO: time based write interval. likely via a leader-triggered signal.
			if s.sinceFlip > s.outstandingLimit {
				s.syncChan <- 1
				s.sinceFlip = 0
			}
		}
	}
}

func (s *Shard) evictOldItems() {
	conf := s.globalConfig.Load().(common.GlobalConfig)
	toRemove := int(float32(int(conf.NumBuckets)*conf.BucketDepth) * conf.LoadFactorStep)
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
	return &cuckoo.Item{int(wa.GlobalSeqNo), wa.Data, int(wa.Bucket1), int(wa.Bucket2)}
}

func (s *Shard) batchRead(req *common.BatchReadRequest) {
	s.log.Trace.Printf("batchRead: enter\n")
	// @todo --- garbage collection
	conf := s.globalConfig.Load().(common.GlobalConfig)
	reply := new(common.BatchReadReply)
	reply.Replies = make([]common.ReadReply, 0, len(req.Args.Args))

	// Run PIR
	reqlength := int(conf.NumBuckets) / 8
	pirvector := make([]byte, reqlength*conf.ReadBatch)
	for batch := 0; batch < len(req.Args.Args); batch += conf.ReadBatch {
		for i := 0; i < conf.ReadBatch; i += 1 {
			offset := batch + i
			reqVector := req.Args.Args[offset].ForTd[0].RequestVector
			//TODO: what's the deal with trust domains? (the forTD parameter)
			copy(pirvector[reqlength*i:reqlength*(i+1)], reqVector)
		}
		responses, err := s.PirServer.Read(pirvector)
		if err != nil {
			s.log.Error.Fatalf("Reading from PIR Server failed: %v", err)
			req.Reply(&common.BatchReadReply{fmt.Sprintf("Failed to read: %v", err), nil})
			return
		}

		replies := make([]common.ReadReply, conf.ReadBatch)
		responseSize := conf.BucketDepth * conf.DataSize
		for i := 0; i < conf.ReadBatch; i += 1 {
			replies[i].Data = responses[i*responseSize : (i+1)*responseSize]
		}
		reply.Replies = append(reply.Replies, replies...)
	}

	// Return results
	req.Reply(reply)
	s.log.Trace.Printf("batchRead: exit\n")
}
