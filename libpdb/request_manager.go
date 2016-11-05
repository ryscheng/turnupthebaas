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
	globalConfig *atomic.Value //*common.GlobalConfig
	dead         int32
	// Channels
	writeChan  chan *common.WriteArgs
	writeQueue []*common.WriteArgs
	readChan   chan *common.ReadArgs
	readQueue  []*common.ReadArgs
}

func NewRequestManager(name string, leader common.LeaderInterface, globalConfig *atomic.Value) *RequestManager {
	rm := &RequestManager{}
	rm.log = common.NewLogger(name)
	rm.leader = leader
	rm.globalConfig = globalConfig
	rm.dead = 0

	rm.log.Info.Printf("NewRequestManager \n")
	go rm.readPeriodic()
	go rm.writePeriodic()
	return rm
}

/** PUBLIC METHODS (threadsafe) **/

func (rm *RequestManager) Kill() {
	atomic.StoreInt32(&rm.dead, 1)
}

func (rm *RequestManager) EnqueueWrite(args *common.WriteArgs) {
	rm.writeChan <- args
}

func (rm *RequestManager) EnqueueRead(args *common.ReadArgs) {
	rm.readChan <- args
}

/** PRIVATE METHODS **/
func (rm *RequestManager) isDead() bool {
	return atomic.LoadInt32(&rm.dead) != 0
}

func (rm *RequestManager) writePeriodic() {
	rand, randErr := drbg.NewHashDrbg(nil)
	if randErr != nil {
		rm.log.Error.Fatalf("Error creating new HashDrbg: %v\n", randErr)
	}

	for rm.isDead() == false {
		select {
		case msg := <-rm.writeChan:
			rm.writeQueue = append(rm.writeQueue, msg)
		default:
			globalConfig := rm.globalConfig.Load().(common.GlobalConfig)
			var args *common.WriteArgs
			var reply common.WriteReply
			if len(rm.writeQueue) > 0 {
				args = rm.writeQueue[0]
				rm.writeQueue = rm.writeQueue[1:]
				rm.log.Info.Printf("writePeriodic: Real request to %v, %v \n", args.Bucket1, args.Bucket2)
			} else {
				args = &common.WriteArgs{}
				rm.generateRandomWrite(globalConfig, rand, args)
				rm.log.Info.Printf("writePeriodic: Dummy request to %v, %v \n", args.Bucket1, args.Bucket2)
			}
			//@todo Do something with response
			startTime := time.Now()
			err := rm.leader.Write(args, &reply)
			elapsedTime := time.Since(startTime)

			if err != nil || reply.Err != "" {
				rm.log.Warn.Printf("writePeriodic error: %v, reply=%v, time=%v\n", err, reply, elapsedTime)
			} else {
				rm.log.Info.Printf("writePeriodic seqNo=%v, time=%v\n", reply.GlobalSeqNo, elapsedTime)
			}
			time.Sleep(globalConfig.WriteInterval)
		}
	}
}

func (rm *RequestManager) readPeriodic() {
	rand, randErr := drbg.NewHashDrbg(nil)
	if randErr != nil {
		rm.log.Error.Fatalf("Error creating new HashDrbg: %v\n", randErr)
	}

	for rm.isDead() == false {
		select {
		case msg := <-rm.readChan:
			rm.readQueue = append(rm.readQueue, msg)
		default:
			globalConfig := rm.globalConfig.Load().(common.GlobalConfig)
			var args *common.ReadArgs
			var reply common.ReadReply
			if len(rm.readQueue) > 0 {
				args = rm.readQueue[0]
				rm.readQueue = rm.readQueue[1:]
				rm.log.Info.Printf("readPeriodic: Real request \n")
			} else {
				args = &common.ReadArgs{}
				rm.generateRandomRead(globalConfig, rand, args)
				rm.log.Info.Printf("readPeriodic: Dummy request \n")
			}
			//@todo Do something with response
			startTime := time.Now()
			err := rm.leader.Read(args, &reply)
			elapsedTime := time.Since(startTime)

			if err != nil || reply.Err != "" {
				rm.log.Error.Printf("readPeriodic error: %v, reply=%v, time=%v\n", err, reply, elapsedTime)
			} else {
				rm.log.Info.Printf("readPeriodic reply: range=%v, time=%v\n", reply.GlobalSeqNo, elapsedTime)
			}
			time.Sleep(globalConfig.ReadInterval)
		}
	}
}

func (rm *RequestManager) generateRandomWrite(globalConfig common.GlobalConfig, rand *drbg.HashDrbg, args *common.WriteArgs) {
	args.Bucket1 = rand.RandomUint64() % globalConfig.NumBuckets
	args.Bucket2 = rand.RandomUint64() % globalConfig.NumBuckets
	args.Data = make([]byte, globalConfig.DataSize, globalConfig.DataSize)
	rand.FillBytes(args.Data)
}

func (rm *RequestManager) generateRandomRead(globalConfig common.GlobalConfig, rand *drbg.HashDrbg, args *common.ReadArgs) {
	numTds := len(globalConfig.TrustDomains)
	numBytes := (uint32(globalConfig.NumBuckets) / uint32(8)) + 1
	if (uint32(globalConfig.NumBuckets) % uint32(8)) > 0 {
		numBytes = numBytes + 1
	}
	args.ForTd = make([]common.PirArgs, numTds, numTds)
	for i := 0; i < numTds; i++ {
		args.ForTd[i].RequestVector = make([]byte, numBytes, numBytes)
		rand.FillBytes(args.ForTd[i].RequestVector)
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
