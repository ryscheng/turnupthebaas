package server

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/foobaz/go-zopfli/zopfli"
	"github.com/privacylab/talek/common"
)

// Frontend terminates client connections to the leader server.
// It is the point of global serialization, and establishes sequence numbers.
type Frontend struct {
	// Private State
	log  *log.Logger
	name string
	*Config

	proposedSeqNo   uint64 // Use atomic.AddUint64, atomic.LoadUint64
	currentInterest *globalInterest
	readChan        chan *readRequest

	replicas []common.ReplicaInterface
	dead     int32

	Verbose bool
}

type globalInterest struct {
	ID               uint64
	CompressedVector []byte
	Signatures       [][32]byte
}

// readRequest is the grouped request and reply memory used for batching
// incoming reads onto a single thread.
type readRequest struct {
	Args  *common.EncodedReadArgs
	Reply *common.ReadReply
	Done  chan bool
}

// NewFrontend creates a new Frontend for a provided configuration.
func NewFrontend(name string, config *Config, replicas []common.ReplicaInterface) *Frontend {
	fe := &Frontend{}
	fe.log = log.New(os.Stdout, "[Frontend:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	fe.name = name
	fe.Config = config
	fe.replicas = replicas
	fe.readChan = make(chan *readRequest, 10)
	nextInterest := new(globalInterest)
	fe.currentInterest = nextInterest

	// Periodically serialize database epoch advances.
	go fe.periodicWrite()
	// Batch incoming reads into combined requests to replicas.
	go fe.batchReads()

	return fe
}

/** PUBLIC METHODS (threadsafe) **/

// Close goroutines associated with this object.
func (fe *Frontend) Close() {
	atomic.StoreInt32(&fe.dead, 1)
}

// GetName exports the name of the server.
func (fe *Frontend) GetName(args *interface{}, reply *string) error {
	*reply = fe.name
	return nil
}

// GetConfig returns the current common configuration from the server.
func (fe *Frontend) GetConfig(args *interface{}, reply *common.Config) error {
	config := *fe.Config.Config
	*reply = config
	return nil
}

func (fe *Frontend) Write(args *common.WriteArgs, reply *common.WriteReply) error {
	seqNo := atomic.AddUint64(&fe.proposedSeqNo, 1)
	args.GlobalSeqNo = seqNo

	replicaWrite := &common.ReplicaWriteArgs{
		WriteArgs: *args,
	}
	replicaReply := common.ReplicaWriteReply{}
	if fe.Verbose {
		fe.log.Printf("write to %d,%d serialized.\n", args.Bucket1, args.Bucket2)
	}
	//@todo writes in parallel
	for i, r := range fe.replicas {
		err := r.Write(replicaWrite, &replicaReply)
		if err != nil {
			reply.Err = err.Error()
			fe.log.Printf("Error writing to replica %d: %v", i, err)
		} else if len(replicaReply.Err) > 0 {
			reply.Err = replicaReply.Err
		}
	}
	reply.GlobalSeqNo = args.GlobalSeqNo

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
	intr := fe.currentInterest
	reply.InterestVector = intr.CompressedVector
	reply.Signature = intr.Signatures
	return nil
}

// periodicWrite runs until the dead flag is set, and periodically send a write
// request to all replicas telling them to advance their write epoch.
func (fe *Frontend) periodicWrite() {
	for atomic.LoadInt32(&fe.dead) == 0 {
		tick := time.After(fe.WriteInterval)
		select {
		case <-tick:
			args := &common.ReplicaWriteArgs{
				EpochFlag: true,
			}
			var rep common.ReplicaWriteReply
			if fe.Verbose {
				fe.log.Printf("Periodic update of database sent to replicas.\n")
			}
			for _, r := range fe.replicas {
				r.Write(args, &rep)
			}
		}
	}
}

func (fe *Frontend) periodicUpdate() {
	// refresh global interest vector from replicas
	for atomic.LoadInt32(&fe.dead) == 0 {
		tick := time.After(time.Duration(fe.WriteInterval.Nanoseconds() * int64(fe.InterestMultiple)))
		select {
		case <-tick:
			args := &common.ReplicaWriteArgs{
				InterestFlag: true,
			}
			resp := make([]common.ReplicaWriteReply, len(fe.replicas))
			if fe.Verbose {
				fe.log.Printf("Periodic update of global interest vector to replicas.\n")
			}
			for i, r := range fe.replicas {
				r.Write(args, &resp[i])
			}
			go fe.generateInterestVector(resp)
		}
	}
}

func (fe *Frontend) generateInterestVector(partials []common.ReplicaWriteReply) {
	// Check that all replicas provided the same interest vector.
	for i, v := range partials {
		if !bytes.Equal(partials[0].InterestVec, v.InterestVec) {
			fe.log.Printf("Replica %d Interest Vector out of sync. Aborting.", i)
			return
		}
	}

	nextInterest := new(globalInterest)
	// TODO: not a bad idea for frontend to validate replica signatures.
	nextInterest.Signatures = make([][32]byte, len(partials))
	for i, v := range partials {
		copy(nextInterest.Signatures[i][:], v.Signature[:])
	}

	// Compress the vector
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	opts := zopfli.DefaultOptions()
	err := zopfli.Compress(
		&opts,
		zopfli.FORMAT_DEFLATE,
		partials[0].InterestVec,
		writer)
	if err != nil {
		fe.log.Printf("Failed to update interest vector delta: %v", err)
		return
	}
	compressed := b.Bytes()
	sum := sha256.Sum256(compressed)

	nextInterest.ID = binary.LittleEndian.Uint64(sum[0:8])
	nextInterest.CompressedVector = compressed
	if fe.Verbose {
		fe.log.Printf("Interest Vector of recent writes regenerated. %d bytes.", len(compressed))
	}
	fe.currentInterest = nextInterest
}

func (fe *Frontend) batchReads() {
	batch := make([]*readRequest, 0, fe.Config.ReadBatch)
	var readReq *readRequest
	tick := time.After(fe.Config.ReadInterval)
	for atomic.LoadInt32(&fe.dead) == 0 {
		select {
		case readReq = <-fe.readChan:
			batch = append(batch, readReq)
			if len(batch) >= fe.Config.ReadBatch {
				go fe.triggerBatchRead(batch)
				batch = make([]*readRequest, 0, fe.Config.ReadBatch)
			} else if fe.Verbose {
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
	if fe.Verbose {
		fe.log.Printf("Batch read with %d items sent to replicas.\n", len(batch))
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
	var replicaErr error
	replies := make([]common.BatchReadReply, len(fe.replicas))
	for i, r := range fe.replicas {
		err := r.BatchRead(args, &replies[i])
		if err != nil || replies[i].Err != "" {
			replicaErr = err
			fe.log.Printf("Error making read to replica %d: %v%v", i, err, replies[i].Err)
			break
		}
		if len(replies[i].Replies) != len(batch) {
			replicaErr = errors.New("failure from Replica " + fmt.Sprintf("%d", i))
			fe.log.Printf("Replica %d gave the wrong number of replies (%d instead of %d)", i, len(replies[i].Replies), len(batch))
			break
		}
	}

	// Respond to clients
	// @todo propagate errors back to clients.
	lastInterestSN := fe.currentInterest.ID
	for i, val := range batch {
		// any error from any replica invalidates the response
		if replicaErr != nil {
			val.Reply.Err = replicaErr.Error()
			val.Done <- true
			break
		}

		replyLength := len(replies[i].Replies[i].Data)
		val.Reply.Data = make([]byte, replyLength)
		for _, rp := range replies {
			val.Reply.Combine(rp.Replies[i].Data)
		}
		val.Reply.GlobalSeqNo = args.SeqNoRange
		val.Reply.LastInterestSN = lastInterestSN
		val.Done <- true
	}

	return nil
}
