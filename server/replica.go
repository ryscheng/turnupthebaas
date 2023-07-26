package server

import (
	"math"
	"math/rand"
	"sync/atomic"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/drbg"
	"github.com/willscott/bloom"
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
	interestVector *bloom.Filter

	// Channels
	ReadBatch []*common.ReadRequest
	ReadChan  chan *common.ReadRequest
	dead      int
	closeChan chan int
}

// NewReplica creates a new Replica server.
func NewReplica(name string, backing string, config Config) *Replica {
	r := &Replica{}
	r.log = common.NewLogger(name)
	r.name = name

	bfSize := math.Ceil(math.Log2(float64(config.NumBuckets)))
	rand := rand.New(rand.NewSource(config.InterestSeed))
	iv, err := bloom.New(rand, int(bfSize), config.BloomFalsePositive)
	if err != nil {
		r.log.Error.Printf("Failed to initialize interest vector: %v", err)
		return nil
	}
	r.interestVector = iv

	r.config.Store(config)

	r.shard = NewShard(name, backing, config)

	return r
}

// Close shuts down active reading and writing threads of the server.
func (r *Replica) Close() {
	// Stop the shard.
	r.shard.Close()
}

/** PUBLIC METHODS (threadsafe) **/
// TODO: need a receive queue serializer to be able to pause on missing seq numbers and
// process only in order - in case conn between leader and replica needs restart.
func (r *Replica) Write(args *common.ReplicaWriteArgs, reply *common.ReplicaWriteReply) error {
	r.log.Trace.Println("Write: enter")
	tr := trace.New("replica.write", "Write")
	defer tr.Finish()

	// update new global interest vector.
	if args.InterestFlag {
		reply.InterestVec = r.interestVector.Delta()
		// TODO: sign
		r.log.Trace.Println("Write-GlobalInterest epoch exit")
		return nil
	}

	r.shard.Write(args)
	r.interestVector.TestAndSet(args.InterestVector)

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
