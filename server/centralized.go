package server

import (
	"sync/atomic"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/drbg"
	"golang.org/x/net/trace"
)

// Centralized talek server implements a replica implemented by a single Centralized
// shard of data.
type Centralized struct {
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

// NewCentralized creates a new Centralized talek server.
func NewCentralized(name string, socket string, config Config) *Centralized {
	c := &Centralized{}
	c.log = common.NewLogger(name)
	c.name = name

	c.config.Store(config)

	c.shard = NewShard(name, socket, config)

	return c
}

// Close shuts down active reading and writing threads of the server.
func (c *Centralized) Close() {
	// Stop the shard.
	c.shard.Close()
}

/** PUBLIC METHODS (threadsafe) **/

func (c *Centralized) Write(args *common.ReplicaWriteArgs, reply *common.ReplicaWriteReply) error {
	c.log.Trace.Println("Write: enter")
	tr := trace.New("centralized.write", "Write")
	defer tr.Finish()

	c.shard.Write(args)

	atomic.StoreUint64(&c.committedSeqNo, args.GlobalSeqNo)
	reply.GlobalSeqNo = args.GlobalSeqNo
	c.log.Trace.Println("Write: exit")
	return nil
}

// BatchRead performs a set of reads against the talek database at one logical point in time.
// BatchRead is replicated to followers with a batching determined by the leader.
func (c *Centralized) BatchRead(args *common.BatchReadRequest, reply *common.BatchReadReply) error {
	c.log.Trace.Println("BatchRead: enter")
	tr := trace.New("centralized.batchread", "BatchRead")
	defer tr.Finish()
	// Start local computation
	config := c.config.Load().(Config)

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
			c.log.Error.Fatalf("Failed to decode part of batch read %v [at index %d]", err, i)
			return err
		}
		localArgs.Args[i] = pir
	}
	c.shard.BatchRead(localArgs)

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
		c.log.Warn.Println("Shard did not respond to all reads!")
		return nil
	}
	reply.Replies = myReply.Replies[0:len(args.Args)]
	c.log.Trace.Println("BatchRead: exit")
	return nil
}

// GetUpdates provies the most recent bloom filter of changed cells.
// TODO
func (c *Centralized) GetUpdates(args *common.GetUpdatesArgs, reply *common.GetUpdatesReply) error {
	c.log.Trace.Println("GetUpdates: ")
	// @TODO
	return nil
}
