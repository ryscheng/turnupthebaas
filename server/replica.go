package server

import (
	"sync/atomic"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/drbg"
	"golang.org/x/net/trace"
)

// Replica implements a centralized talek replica.
type Replica struct {
	/** Private State **/
	// Static
	log    *common.Logger
	name   string
	status chan int

	// Thread-safe
	config         atomic.Value //Config
	shard          *Shard
	committedSeqNo uint64 // Use atomic.AddUint64, atomic.LoadUint64
	// Channels
	ReadBatch []*common.ReadRequest
	ReadChan  chan *common.ReadRequest
	dead      int
	closeChan chan int
}

// NewReplica creates a new Replica server.
func NewReplica(name string, socket string, config Config) *Replica {
	r := &Replica{}
	r.log = common.NewLogger(name)
	r.name = name

	r.config.Store(config)

	r.shard = NewShard(name, socket, config)

	return r
}

// Close shuts down active reading and writing threads of the server.
func (r *Replica) Close() {
	// Stop the shard.
	r.shard.Close()
}

/** PUBLIC METHODS (threadsafe) **/

func (r *Replica) Write(args *common.ReplicaWriteArgs, reply *common.ReplicaWriteReply) error {
	r.log.Trace.Println("Write: enter")
	tr := trace.New("replica.write", "Write")
	defer tr.Finish()

	// update new global interest vector.
	if args.InterestFlag {
		reply.InterestVec = nil // new interest vector.
	}

	r.shard.Write(args)

	atomic.StoreUint64(&r.committedSeqNo, args.GlobalSeqNo)
	reply.GlobalSeqNo = args.GlobalSeqNo
	r.log.Trace.Println("Write: exit")
	return nil
}

// BatchRead performs a set of reads against the talek database at one logical point in time.
// BatchRead is replicated to followers with a batching determined by the leader.
func (r *Replica) BatchRead(args *common.BatchReadRequest, reply *common.BatchReadReply) error {
	r.log.Trace.Println("BatchRead: enter")
	tr := trace.New("replica.batchread", "BatchRead")
	defer tr.Finish()
	// Start local computation
	config := r.config.Load().(Config)

	localArgs := new(DecodedBatchReadRequest)
	localArgs.ReplyChan = make(chan *common.BatchReadReply)
	localArgs.Args = make([]common.PirArgs, config.ReadBatch)
	for i, val := range args.Args {
		//Handle pad requests.
		if len(val.PirArgs) == 0 {
			localArgs.Args[i].PadSeed = make([]byte, drbg.SeedLength)
			localArgs.Args[i].RequestVector = make([]byte, config.NumBuckets/8)
			continue
		}
		pir, err := val.Decode(config.TrustDomainIndex, config.TrustDomain)
		if err != nil {
			reply.Err = err.Error()
			r.log.Error.Fatalf("Failed to decode part of batch read %v [at index %d]", err, i)
			return err
		}
		localArgs.Args[i] = pir
	}
	r.shard.BatchRead(localArgs)

	// wait for results
	myReply := <-localArgs.ReplyChan

	// Mutate results
	for i, val := range localArgs.Args {
		if myReply.Replies[i].Err == "" {
			if err := drbg.Overlay(val.PadSeed, myReply.Replies[i].Data); err != nil {
				myReply.Replies[i].Err = err.Error()
			}
		}
	}

	if len(args.Args) > len(myReply.Replies) {
		r.log.Warn.Println("Shard did not respond to all reads!")
		return nil
	}
	reply.Replies = myReply.Replies[0:len(args.Args)]
	r.log.Trace.Println("BatchRead: exit")
	return nil
}

// GetUpdates returns the signed version of the current global interest vector
func (r *Replica) GetUpdates(args *common.ReplicaUpdateArgs, reply *common.ReplicaUpdateReply) error {
	r.log.Trace.Println("GetUpdates: enter")
	tr := trace.New("replica.getupdates", "GetUpdates")
	defer tr.Finish()

	r.log.Trace.Println("GetUpdates: exit")
	return nil
}
