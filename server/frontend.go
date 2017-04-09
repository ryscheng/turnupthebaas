package server

import (
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/privacylab/talek/common"
)

// Frontend terminates client connections to the leader server.
// It is the point of global serialization, and establishes sequence numbers.
type Frontend struct {
	// Private State
	log  *log.Logger
	name string
	*Config

	proposedSeqNo uint64 // Use atomic.AddUint64, atomic.LoadUint64
	readChan      chan *readRequest

	replicas []common.ReplicaInterface
	dead     int
}

// readRequest is the grouped request and reply memory used for batching
// incoming reads onto a single thread.
type readRequest struct {
	Args  *common.EncodedReadArgs
	Reply *common.ReadReply
	Done  chan bool
}

// NewFrontend creates a new Frontend for a provided configuration.
func NewFrontend(name string, serverConfig *Config, replicas []common.ReplicaInterface) *Frontend {
	fe := &Frontend{}
	fe.log = log.New(os.Stdout, "[Frontend:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	fe.name = name
	fe.Config = serverConfig
	fe.replicas = replicas
	fe.readChan = make(chan *readRequest, 10)

	// Periodically serialize database epoch advances.
	go fe.periodicWrite()
	// Batch incoming reads into combined requests to replicas.
	go fe.batchReads()

	return fe
}

/** PUBLIC METHODS (threadsafe) **/

// Close goroutines associated with this object.
func (fe *Frontend) Close() {
	fe.dead = 1
}

// GetName exports the name of the server.
func (fe *Frontend) GetName(args *interface{}, reply *string) error {
	*reply = fe.name
	return nil
}

// GetConfig returns the current common configuration from the server.
func (fe *Frontend) GetConfig(args *interface{}, reply *common.Config) error {
	config := fe.Config
	*reply = *config.Config
	return nil
}

func (fe *Frontend) Write(args *common.WriteArgs, reply *common.WriteReply) error {
	seqNo := atomic.AddUint64(&fe.proposedSeqNo, 1)
	args.GlobalSeqNo = seqNo

	replicaWrite := &common.ReplicaWriteArgs{
		WriteArgs: *args,
	}
	replicaReply := common.ReplicaWriteReply{}
	//@todo writes in parallel
	for i, r := range fe.replicas {
		err := r.Write(replicaWrite, &replicaReply)
		if err != nil {
			reply.Err = err.Error()
			fe.log.Fatalf("Error writing to replica %d: %v", i, err)
		} else if len(replicaReply.Err) > 0 {
			reply.Err = replicaReply.Err
		}
	}

	return nil
}

func (fe *Frontend) Read(args *common.EncodedReadArgs, reply *common.ReadReply) error {
	ready := make(chan bool, 1)
	fe.readChan <- &readRequest{Args: args, Reply: reply, Done: ready}
	<-ready

	return nil
}

// GetUpdates provides the most recent global interest vector deltas.
func (fe *Frontend) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	fe.log.Println("GetUpdates: ")
	// @TODO
	return nil
}

// periodicWrite runs until the dead flag is set, and periodically send a write
// request to all replicas telling them to advance their write epoch.
func (fe *Frontend) periodicWrite() {
	for fe.dead == 0 {
		tick := time.After(fe.Config.WriteInterval)
		select {
		case <-tick:
			args := &common.ReplicaWriteArgs{
				EpochFlag: true,
			}
			var rep common.ReplicaWriteReply
			for _, r := range fe.replicas {
				r.Write(args, &rep)
			}
		}
	}
}

func (fe *Frontend) batchReads() {
	batch := make([]*readRequest, 0, fe.Config.ReadBatch)
	var readReq *readRequest
	tick := time.After(fe.Config.ReadInterval)
	for fe.dead == 0 {
		select {
		case readReq = <-fe.readChan:
			batch = append(batch, readReq)
			if len(batch) >= fe.Config.ReadBatch {
				go fe.triggerBatchRead(batch)
				batch = make([]*readRequest, 0, fe.Config.ReadBatch)
			} else {
				fe.log.Printf("Read: add to batch, size=%v\n", len(batch))
			}
			continue
		case <-tick:
			if len(batch) > 0 {
				go fe.triggerBatchRead(batch)
				batch = make([]*readRequest, 0, fe.Config.ReadBatch)
			}
			tick = time.After(fe.Config.ReadInterval)
			continue
		}
	}
}

func (fe *Frontend) triggerBatchRead(batch []*readRequest) error {
	args := &common.BatchReadRequest{}
	// Copy args
	args.Args = make([]common.EncodedReadArgs, len(batch), len(batch))
	for i, val := range batch {
		if val.Args != nil {
			args.Args[i] = *val.Args
		}
	}

	// Choose a SeqNoRange
	currSeqNo := atomic.LoadUint64(&fe.proposedSeqNo) + 1
	if currSeqNo <= uint64(fe.Config.WindowSize()) {
		args.SeqNoRange.Start = 1 // Minimum of 1
	} else {
		args.SeqNoRange.Start = currSeqNo - uint64(fe.Config.WindowSize()) // Inclusive
	}
	args.SeqNoRange.End = currSeqNo // Exclusive
	args.SeqNoRange.Aborted = make([]uint64, 0, 0)

	// Start computation
	// @todo reads in parallel
	replies := make([]common.BatchReadReply, len(fe.replicas))
	for i, r := range fe.replicas {
		err := r.BatchRead(args, &replies[i])
		if err != nil || replies[i].Err != "" {
			fe.log.Fatalf("Error making read to replica %d: %v%v", i, err, replies[i].Err)
		}
		if len(replies[i].Replies) != len(batch) {
			fe.log.Fatalf("Replica %d gave the wrong number of replies (%d instead of %d)", i, len(replies[i].Replies), len(batch))
		}
	}

	// Respond to clients
	// @todo propagate errors back to clients.
	for i, val := range batch {
		for _, rp := range replies {
			val.Reply.Combine(rp.Replies[i].Data)
		}
		val.Reply.GlobalSeqNo = args.SeqNoRange
		val.Done <- true
	}

	return nil
}
