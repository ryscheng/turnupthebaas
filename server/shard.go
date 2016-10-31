package server

import (
	"fmt"
	"github.com/ryscheng/pdb/common"
	"github.com/ryscheng/pdb/cuckoo"
	"github.com/ryscheng/pdb/pir"
	"log"
	"os"
	"sync/atomic"
)

const maxWriteBuffer int = 1024

/**
 * Handles a shard of the data
 * Goroutines:
 * - 1x processRequests()
 */
type Shard struct {
	// Private State
	log           *log.Logger
	name          string
	WriteLog      map[uint64]*common.WriteArgs
	pendingWrites []uint64
	globalConfig  atomic.Value //common.GlobalConfig
	pir.PirServer

	// Channels
	WriteChan     chan *common.WriteArgs
	BatchReadChan chan *BatchReadRequest
}

func NewShard(name string, globalConfig common.GlobalConfig) *Shard {
	s := &Shard{}
	s.log = log.New(os.Stdout, "["+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	s.name = name
	s.pendingWrites = make([]uint64, 0, maxWriteBuffer)
	s.WriteLog = make(map[uint64]*common.WriteArgs)
	s.globalConfig.Store(globalConfig)
	s.WriteChan = make(chan *common.WriteArgs)
	s.BatchReadChan = make(chan *BatchReadRequest)

	// TODO: per-server config of where the local PIR socket is.
	s.PirServer, err = pir.Connect("pir.socket")
	if err != nil {
		s.log.Fatalf("Could not connect to pir back end: %v", err)
		return nil
	}
	err = s.PirServer.Configure(globalConfig.DataSize*globalConfig.BucketDepth, globalConfig.NumBuckets, globalConfig.ReadBatch)
	if err != nil {
		s.log.Fatalf("Could not start PIR back end with correct parameters: %v", err)
		return nil
	}
	s.PirDB, err = s.PirServer.GetDB()
	if err != nil {
		s.log.Fatalf("Could not allocate memory for inital server database: %v", err)
		return nil
	}

	go s.processRequests()
	return s
}

/** PUBLIC METHODS (threadsafe) **/
func (s *Shard) Ping(args *common.PingArgs, reply *common.PingReply) error {
	s.log.Println("Ping: " + args.Msg + ", ... Pong")
	reply.Err = ""
	reply.Msg = "PONG"
	return nil
}

func (s *Shard) Write(args *common.WriteArgs, reply *common.WriteReply) error {
	//s.log.Println("Write: ")
	s.WriteChan <- args
	reply.Err = ""
	return nil
}

func (s *Shard) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	s.log.Println("GetUpdates: ")
	// @TODO
	reply.Err = ""
	//reply.InterestVector =
	return nil
}

func (s *Shard) BatchRead(args *common.BatchReadArgs, replyChan chan *common.BatchReadReply) error {
	s.log.Println("Read: ")
	batchReq := &BatchReadRequest{args, replyChan}
	s.BatchReadChan <- batchReq
	return nil
}

/** PRIVATE METHODS (singlethreaded) **/
func (s *Shard) processRequests() {
	var writeReq *common.WriteArgs
	var batchReadReq *BatchReadRequest

	defer s.PirServer.Disconnect()
	for {
		select {
		case writeReq = <-s.WriteChan:
			s.processWrite(writeReq)
		case batchReadReq = <-s.BatchReadChan:
			s.batchRead(batchReadReq)
		}
	}
}

func asCuckooItem(wa *common.WriteArgs) cuckoo.Item {
	return &cuckoo.Item{wa.Data, wa.Bucket1, wa.Bucket2}
}

func (s *Shard) processWrite(req *common.WriteArgs) {
	s.log.Printf("processWrite: seqNo=%v\n", req.GlobalSeqNo)
	s.WriteLog[req.GlobalSeqNo] = req
	//s.log.Printf("%v\n", s.WriteLog)
	append(s.pendingWrites, req.GlobalSeqNo)

	// Trigger to swap to next DB.
	if len(s.pendingWrites) == maxWriteBuffer {
		conf := s.globalConfig.Load().(common.GlobalConfig)
		// TODO: check if globalConfig has changed / server needs to be fully reconfigured.
		newDB, err := s.PirServer.GetDB()
		if err != nil {
			s.log.Fatalf("Could not update shared. Failed to allocate memory for DB update: %v", err)
			panic("Could not update shared. Failed to allocate memory for DB update")
		}

		// TODO: where does random seed come from?
		randSeed := 0
		table := cuckoo.NewTable(s.name, conf.NumBuckets, conf.BucketDepth, conf.DataSize, newDB.DB, randSeed)
		// TODO: does table start from previous snapshto, or does it explicitly insert the full log of
		// writes on each snapshot?
		// TODO: how would old items fall out of table?
		for i := 0; i < len(s.pendingWrites); i++ {
			table.Insert(asCuckooItem(s.WriteLog[s.pendingWrites[i]]))
		}

		err = s.PirServer.SetDB(newDB)
		if err != nil {
			s.log.Fatalf("Failed to swap DB to new DB: %v", err)
			panic("Could not update to next DB snapshot.")
		}
		if s.PirDB != nil {
			s.PirDB.Free()
		}
		s.PirDB = newDB
		s.pendingWrites = s.pendingWrites[0:0]
	}
}

func (s *Shard) batchRead(req *BatchReadRequest) {
	// @todo --- garbage collection
	conf := s.globalConfig.Load().(common.GlobalConfig)
	reply := new(common.BatchReadReply)
	reply.Replies = make(common.ReadReply, 0, len(req.Args.Args))

	// Run PIR
	pirvector := make([]byte, conf.NumBuckets/8*conf.ReadBatch)
	for batch := 0; batch < len(req.Args.Args); batch += conf.ReadBatch {
		for i := 0; i < conf.ReadBatch; i += 1 {
			offset := batch*conf.ReadBatch + i
			copy(pirvector[conf.NumBuckets/8*i:], req.Args.Args[offset].RequestVector)
		}
		responses, err := s.PirServer.Read(pirvector)
		if err != nil {
			s.log.Fatalf("Reading from PIR Server failed: %v", err)
			req.Reply(&common.BatchReadReply{fmt.Sprintf("Failed to read: %v", err)})
			return
		}

		replies := make([]common.ReadReply, conf.ReadBatch)
		responseSize := conf.BucketDepth * conf.DataSize
		for i := 0; i < conf.ReadBatch; i += 1 {
			offset := i * conf.BucketDepth * conf.DataSize
			replies[i].Data = responses[i*responseSize : (i+1)*responseSize]
		}
		append(reply.Replies, replies...)
	}

	// Return results
	req.Reply(reply)
}
