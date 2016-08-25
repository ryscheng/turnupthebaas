package libpdb

import (
	"github.com/ryscheng/pdb/common"
	"github.com/ryscheng/pdb/drbg"
	"log"
	"os"
	"sync/atomic"
	"time"
)

//const defaultReadInterval = int64(time.Second)
//const defaultWriteInterval = int64(time.Second)

type RequestManager struct {
	log       *log.Logger
	serverRef *common.TrustDomainRef
	// Protected by `atomic`
	globalConfig *atomic.Value //*common.GlobalConfig
	dead         int32
}

func NewRequestManager(name string, serverRef *common.TrustDomainRef, globalConfig *atomic.Value) *RequestManager {
	rm := &RequestManager{}
	rm.log = log.New(os.Stdout, "[RequestManager:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	rm.serverRef = serverRef
	rm.globalConfig = globalConfig
	rm.dead = 0

	rm.log.Printf("NewRequestManager \n")
	//go rm.readPeriodic()
	go rm.writePeriodic()
	return rm
}

/** PUBLIC METHODS (threadsafe) **/

func (rm *RequestManager) Kill() {
	atomic.StoreInt32(&rm.dead, 1)
}

/** PRIVATE METHODS **/
func (rm *RequestManager) isDead() bool {
	return atomic.LoadInt32(&rm.dead) != 0
}

func (rm *RequestManager) writePeriodic() {
	seed, _ := drbg.NewSeed()
	rand := drbg.NewHashDrbg(seed)
	for rm.isDead() == false {
		// Load latest config
		globalConfig := rm.globalConfig.Load().(common.GlobalConfig)
		args := &common.WriteArgs{}
		rm.generateRandomWrite(globalConfig, rand, args)
		rm.log.Printf("writePeriodic: Dummy request to %v, %v \n", args.Bucket1, args.Bucket2)
		time.Sleep(globalConfig.WriteInterval)
		//time.Sleep(time.Duration(atomic.LoadInt64(&rm.writeInterval)))
	}
}

func (rm *RequestManager) readPeriodic() {
	seed, _ := drbg.NewSeed()
	rand := drbg.NewHashDrbg(seed)
	for rm.isDead() == false {
		rm.log.Println("readPeriodic: ")
		globalConfig := rm.globalConfig.Load().(common.GlobalConfig)
		args := &common.ReadArgs{}
		rm.generateRandomRead(globalConfig, rand, args)
		rm.log.Printf("readPeriodic: Dummy request \n")
		time.Sleep(globalConfig.ReadInterval)
		//time.Sleep(time.Duration(atomic.LoadInt64(&rm.readInterval)))
	}
}

func (rm *RequestManager) generateRandomWrite(globalConfig common.GlobalConfig, rand *drbg.HashDrbg, args *common.WriteArgs) {
	args.Bucket1 = rand.RandomUint32() % globalConfig.NumBuckets
	args.Bucket2 = rand.RandomUint32() % globalConfig.NumBuckets
	args.Data = make([]byte, globalConfig.DataSize, globalConfig.DataSize)
	rand.FillBytes(&args.Data)
}

func (rm *RequestManager) generateRandomRead(globalConfig common.GlobalConfig, rand *drbg.HashDrbg, args *common.ReadArgs) {
	numBytes := (globalConfig.NumBuckets / uint32(8))
	lastNumBits := (globalConfig.NumBuckets % uint32(8))
	if lastNumBits > 0 {
		numBytes = numBytes + 1
	}
	args.RequestVector = make([]byte, numBytes, numBytes)
	rand.FillBytes(&args.RequestVector)
	//@todo Trim last byte to expected number of bits

}
