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
	*pir.PirServer
	*Table

	// Channels
	writeChan     chan *common.WriteArgs
	readChan chan *BatchReadRequest
	signalChan    chan int
  sinceFlip int
}

func NewShard(name string, socket string, globalConfig common.GlobalConfig) *Shard {
	s := &Shard{}
	s.log = log.New(os.Stdout, "["+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	s.name = name
	s.WriteLog = make(map[uint64]*common.WriteArgs)
	s.globalConfig.Store(globalConfig)
	s.writeChan = make(chan *common.WriteArgs)
	s.readChan = make(chan *BatchReadRequest)
	s.signalChan = make(chan int)

	// TODO: per-server config of where the local PIR socket is.
	pirServer, err := pir.Connect(socket)
	if err != nil {
		s.log.Fatalf("Could not connect to pir back end: %v", err)
		return nil
	}
	s.PirServer = pirServer
	err = s.PirServer.Configure(globalConfig.DataSize*globalConfig.BucketDepth, int(globalConfig.NumBuckets), globalConfig.ReadBatch)
	if err != nil {
		s.log.Fatalf("Could not start PIR back end with correct parameters: %v", err)
		return nil
	}
	s.Table = NewTable(s.PirServer, name + "-Table", s.log, globalConfig.BucketDepth, globalConfig.MaxLoadFactor, globalConfig.LoadFactorStep)

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
	s.writeChan <- args
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
	s.readChan <- batchReq
	return nil
}

func (s *Shard) Close() {
	s.log.Println("Close: ")
	s.signalChan <- 1
}

/** PRIVATE METHODS (singlethreaded) **/
func (s *Shard) processRequests() {
	var writeReq *common.WriteArgs
	var batchReadReq *BatchReadRequest

	defer s.Table.Close()
	defer s.PirServer.Disconnect()
	for {
		select {
		case writeReq = <-s.writeChan:
			if err := s.processWrite(writeReq); err != nil {
				break
			}
			continue
		case batchReadReq = <-s.readChan:
			s.batchRead(batchReadReq)
			continue
		case <- s.signalChan:
			s.log.Printf("Shard Closing.")
			break
		}
	}
}

func asCuckooItem(wa *common.WriteArgs) *cuckoo.Item {
	//TODO: cuckoo should continue int64 sized buckets if needed.
	return &cuckoo.Item{wa.Data, int(wa.Bucket1), int(wa.Bucket2)}
}

func (s *Shard) processWrite(req *common.WriteArgs) error {
	s.log.Printf("processWrite: seqNo=%v\n", req.GlobalSeqNo)

	err := s.Table.Write(asCuckooItem(req))
	if err != nil {
		s.log.Fatalf("Could not write item: %v", err)
		return err
	}
	s.sinceFlip +=1

	// Trigger to swap to next DB.
	if s.sinceFlip > maxWriteBuffer {
		err := s.Table.Flop()
		s.sinceFlip = 0
		if err != nil {
			s.log.Fatalf("Could not flip write snapshot: %v", err)
			return err
		}
	}
	return nil
}

func (s *Shard) batchRead(req *BatchReadRequest) {
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
			//TODO: what's the deal with trust domains? (the forTD parameter)
			copy(pirvector[reqlength * i:reqlength*(i+1)], req.Args.Args[offset].ForTd[0].RequestVector)
		}
		responses, err := s.PirServer.Read(pirvector)
		if err != nil {
			s.log.Fatalf("Reading from PIR Server failed: %v", err)
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
}
