package libpdb

import (
	"github.com/ryscheng/pdb/common"
	"github.com/ryscheng/pdb/drbg"
	"sync/atomic"
	"time"
)

//const defaultReadInterval = int64(time.Second)
//const defaultWriteInterval = int64(time.Second)

type RequestManager struct {
	log    *common.Logger
	leader common.LeaderInterface
	// Protected by `atomic`
	config *atomic.Value //*ClientConfig
	dead         int32
	rand         *drbg.HashDrbg
	// Channels
	writeChan  chan *common.WriteArgs
	writeQueue []*common.WriteArgs
	readChan   chan *common.ReadRequest
	readQueue  []*common.ReadRequest
}

func NewRequestManager(name string, leader common.LeaderInterface, config *atomic.Value) *RequestManager {
	rm := &RequestManager{}
	rm.log = common.NewLogger(name)
	rm.leader = leader
	rm.config = config
	rm.dead = 0

	rand, randErr := drbg.NewHashDrbg(nil)
	if randErr != nil {
		rm.log.Error.Fatalf("Error creating new HashDrbg: %v\n", randErr)
	}
	rm.rand = rand
	rm.writeChan = make(chan *common.WriteArgs)
	rm.writeQueue = make([]*common.WriteArgs, 0)
	rm.readChan = make(chan *common.ReadRequest)
	rm.readQueue = make([]*common.ReadRequest, 0)

	rm.log.Info.Printf("NewRequestManager \n")
	go rm.readPeriodic()
	go rm.writePeriodic()
	return rm
}

/** PUBLIC METHODS (threadsafe) **/

func (rm *RequestManager) Kill() {
	atomic.StoreInt32(&rm.dead, 1)
}

func (rm *RequestManager) EnqueueWrite(req *common.WriteArgs) {
	rm.writeChan <- req
}

func (rm *RequestManager) EnqueueRead(req *common.ReadRequest) {
	rm.readChan <- req
}

/** PRIVATE METHODS **/
func (rm *RequestManager) isDead() bool {
	return atomic.LoadInt32(&rm.dead) != 0
}

func (rm *RequestManager) writePeriodic() {

	for rm.isDead() == false {
		select {
		case msg := <-rm.writeChan:
			rm.log.Trace.Println("EnqueueWrite to writeQueue")
			rm.writeQueue = append(rm.writeQueue, msg)
		default:
			config := rm.config.Load().(ClientConfig)
			var req *common.WriteArgs = nil
			var reply common.WriteReply
			if len(rm.writeQueue) > 0 {
				req = rm.writeQueue[0]
				rm.writeQueue = rm.writeQueue[1:]
				rm.log.Info.Printf("writePeriodic: Real request to %v, %v \n", req.Bucket1, req.Bucket2)
			} else {
				req = &common.WriteArgs{}
				rm.generateRandomWrite(config, req)
				rm.log.Info.Printf("writePeriodic: Dummy request to %v, %v \n", req.Bucket1, req.Bucket2)
			}
			//@todo Do something with response
			startTime := time.Now()
			err := rm.leader.Write(req, &reply)
			elapsedTime := time.Since(startTime)

			if err != nil {
				reply.Err = err.Error()
			}
			if reply.Err != "" {
				rm.log.Warn.Printf("writePeriodic error: %v, reply=%v, time=%v\n", err, reply, elapsedTime)
			} else {
				rm.log.Info.Printf("writePeriodic seqNo=%v, time=%v\n", reply.GlobalSeqNo, elapsedTime)
			}

			if req != nil {
				req.ReplyChan <- &reply
			}
			time.Sleep(config.WriteInterval)
		}
	}
}

func (rm *RequestManager) readPeriodic() {

	for rm.isDead() == false {
		select {
		case msg := <-rm.readChan:
			rm.readQueue = append(rm.readQueue, msg)
		default:
			config := rm.config.Load().(ClientConfig)
			var req *common.ReadRequest = nil
			var args *common.ReadArgs
			var reply common.ReadReply
			if len(rm.readQueue) > 0 {
				req = rm.readQueue[0]
				args = req.Args
				rm.readQueue = rm.readQueue[1:]
				rm.log.Info.Printf("readPeriodic: Real request \n")
			} else {
				args = &common.ReadArgs{}
				rm.generateRandomRead(config, args)
				rm.log.Info.Printf("readPeriodic: Dummy request \n")
			}
			//@todo Do something with response
			startTime := time.Now()
			err := rm.leader.Read(args, &reply)
			elapsedTime := time.Since(startTime)

			if err != nil {
				reply.Err = err.Error()
			}
			if reply.Err != "" {
				rm.log.Error.Printf("readPeriodic error: %v, reply=%v, time=%v\n", err, reply, elapsedTime)
			} else {
				rm.log.Info.Printf("readPeriodic reply: range=%v, time=%v\n", reply.GlobalSeqNo, elapsedTime)
			}

			if req != nil {
				req.Reply(&reply)
			}
			time.Sleep(config.ReadInterval)
		}
	}
}

func (rm *RequestManager) generateRandomWrite(config ClientConfig, args *common.WriteArgs) {
	args.Bucket1 = rm.rand.RandomUint64() % config.CommonConfig.NumBuckets
	args.Bucket2 = rm.rand.RandomUint64() % config.CommonConfig.NumBuckets
	args.Data = make([]byte, config.CommonConfig.DataSize, config.CommonConfig.DataSize)
	rm.rand.FillBytes(args.Data)
}

func (rm *RequestManager) generateRandomRead(config ClientConfig, args *common.ReadArgs) {
	numTds := len(config.TrustDomains)
	numBytes := (uint32(config.CommonConfig.NumBuckets) / uint32(8)) + 1
	if (uint32(config.CommonConfig.NumBuckets) % uint32(8)) > 0 {
		numBytes = numBytes + 1
	}
	args.ForTd = make([]common.PirArgs, numTds, numTds)
	for i := 0; i < numTds; i++ {
		args.ForTd[i].RequestVector = make([]byte, numBytes, numBytes)
		rm.rand.FillBytes(args.ForTd[i].RequestVector)
		seed, seedErr := drbg.NewSeed()
		if seedErr != nil {
			rm.log.Error.Fatalf("Error creating new Seed: %v\n", seedErr)
		}
		args.ForTd[i].PadSeed = seed.Export()
	}
	//args.RequestVector = make([]byte, numBytes, numBytes)
	//rand.FillBytes(&args.RequestVector)
	//@todo Trim last byte to expected number of bits

}
