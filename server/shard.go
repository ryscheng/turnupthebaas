package server

import (
	"github.com/ryscheng/pdb/common"
	"log"
	"os"
)

/**
 * Handles a shard of the data
 * Goroutines:
 * - 1x processRequests()
 */
type Shard struct {
	// Private State
	log       *log.Logger
	name      string
	WriteLog  map[uint64]*common.WriteArgs
	ReadBatch []*ReadRequest
	// Channels
	WriteChan    chan *common.WriteArgs
	ReadChan     chan *ReadRequest
	TriggerBatch chan *common.Range
}

func NewShard(name string) *Shard {
	s := &Shard{}
	s.log = log.New(os.Stdout, "[Shard:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	s.name = name
	s.WriteLog = make(map[uint64]*common.WriteArgs)
	s.ReadBatch = make([]*ReadRequest, 0)
	s.WriteChan = make(chan *common.WriteArgs)
	s.ReadChan = make(chan *ReadRequest)

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
	s.log.Println("Write: ")
	s.WriteChan <- args
	reply.Err = ""
	return nil
}

func (s *Shard) Read(args *common.ReadArgs, reply *common.ReadReply) error {
	s.log.Println("Read: ")
	resultChan := make(chan []byte)
	s.ReadChan <- &ReadRequest{args, resultChan}
	reply.Err = ""
	reply.Data = <-resultChan
	return nil
}

func (s *Shard) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	s.log.Println("GetUpdates: ")
	// @TODO
	reply.Err = ""
	//reply.InterestVector =
	return nil
}

/** PRIVATE METHODS (singlethreaded) **/
func (s *Shard) processRequests() {
	var writeReq *common.WriteArgs
	var readReq *ReadRequest
	var dataRange *common.Range
	for {
		select {
		case writeReq = <-s.WriteChan:
			s.processWrite(writeReq)
		case readReq = <-s.ReadChan:
			s.processRead(readReq)
		case dataRange = <-s.TriggerBatch:
			s.batchRead(dataRange)
		}
	}
}

func (s *Shard) processWrite(req *common.WriteArgs) {
	s.WriteLog[req.GlobalSeqNo] = req
}

func (s *Shard) processRead(req *ReadRequest) {
	s.ReadBatch = append(s.ReadBatch, req)
}

func (s *Shard) batchRead(dataRange *common.Range) {
	// build a database
	for len(s.ReadBatch) > 0 {
		// Take batch size and PIR it

	}
}
