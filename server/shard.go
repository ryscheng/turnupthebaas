package server

import (
	"github.com/ryscheng/pdb/common"
	//	"github.com/ryscheng/pdb/cuckoo"
	"log"
	"os"
	"sync/atomic"
)

/**
 * Handles a shard of the data
 * Goroutines:
 * - 1x processRequests()
 */
type Shard struct {
	// Private State
	log          *log.Logger
	name         string
	WriteLog     map[uint64]*common.WriteArgs
	globalConfig atomic.Value //common.GlobalConfig
	// Channels
	WriteChan     chan *common.WriteArgs
	BatchReadChan chan *BatchReadRequest
}

func NewShard(name string, globalConfig common.GlobalConfig) *Shard {
	s := &Shard{}
	s.log = log.New(os.Stdout, "["+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	s.name = name
	s.WriteLog = make(map[uint64]*common.WriteArgs)
	s.globalConfig.Store(globalConfig)
	s.WriteChan = make(chan *common.WriteArgs)
	s.BatchReadChan = make(chan *BatchReadRequest)

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
	for {
		select {
		case writeReq = <-s.WriteChan:
			s.processWrite(writeReq)
		case batchReadReq = <-s.BatchReadChan:
			s.batchRead(batchReadReq)
		}
	}
}

func (s *Shard) processWrite(req *common.WriteArgs) {
	s.log.Printf("processWrite: seqNo=%v\n", req.GlobalSeqNo)
	s.WriteLog[req.GlobalSeqNo] = req
	//s.log.Printf("%v\n", s.WriteLog)
}

func (s *Shard) batchRead(req *BatchReadRequest) {
	// @todo --- garbage collection
	//globalConfig := s.globalConfig.Load().(common.GlobalConfig)
	//table := cuckoo.NewTable(s.name, globalConfig.NumBuckets, globalConfig.BucketDepth, req.Args.RandSeed)

	// build a database
	//for len(s.ReadBatch) > 0 {
	// Take batch size and PIR it

	//}

	// Run PIR over database

	// Return results
}
