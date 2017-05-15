//+build !nocuda,!travis

package pircuda

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unsafe"

	"github.com/barnex/cuda5/cu"
	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/pir/pirinterface"
)

// ShardCUDA represents a read-only shard of the database,
// backed by a CUDA implementation of PIR
type ShardCUDA struct {
	log        *common.Logger
	name       string
	context    *ContextCUDA
	bucketSize int
	numBuckets int
	data       []byte
	numThreads int
	cudaData   cu.DevicePtr
}

// NewShard creates a new cuda shard conforming to the common interface
func NewShard(bucketSize int, data []byte, userdata string) pirinterface.Shard {
	parts := strings.Split(userdata, ".")
	if len(parts) < 4 {
		fmt.Fprintf(os.Stderr, "Invalid cuda specification: %s. Should be cuda.[context].[contextdatasize].[threads]", parts)
		return nil
	}

	contextDataSize, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid datasize: %s. Should be numeric", parts[2])
		return nil
	}

	context, err := NewContextCUDA("contextcuda", parts[1], int(contextDataSize))
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot create new ContextCUDA: error=%v\n", err)
	}

	threads, err := strconv.ParseInt(parts[3], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid threads: %s. Should be numeric", parts[3])
		return nil
	}
	shard, err := NewShardCUDA("CUDA Shard ("+userdata+")", context, bucketSize, data, int(threads))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create cuda shard: %v", err)
		return nil
	}
	return pirinterface.Shard(shard)
}

func init() {
	pirinterface.Register("cuda", NewShard)
}

// NewShardCUDA creates a new CUDA-backed shard
// The data is represented as a flat byte array = append(bucket_1, bucket_2 ... bucket_n)
// Pre-conditions:
// - len(data) must be a multiple of bucketSize
// Returns: the shard, or an error if mismatched size
func NewShardCUDA(name string, context *ContextCUDA, bucketSize int, data []byte, numThreads int) (*ShardCUDA, error) {
	s := &ShardCUDA{}
	s.log = common.NewLogger(name)
	s.name = name
	s.context = context

	// GetNumBuckets will compute the number of buckets stored in the Shard
	// If len(s.data) is not cleanly divisible by s.bucketSize,
	// returns an error
	if len(data)%bucketSize != 0 {
		rErr := fmt.Errorf("NewShardCUDA(%v) failed: data(len=%v) not multiple of bucketSize=%v", name, len(data), bucketSize)
		s.log.Error.Printf("%v\n", rErr)
		return nil, rErr
	}

	s.bucketSize = bucketSize
	s.numBuckets = (len(data) / bucketSize)
	s.data = data
	s.numThreads = numThreads

	/** CUDA **/
	// Weird context hack
	if cu.CtxGetCurrent() == 0 {
		s.context.Ctx.SetCurrent()
	}
	//  Create buffers
	s.cudaData = cu.MemAlloc(int64(len(data)))
	cu.MemcpyHtoD(s.cudaData, unsafe.Pointer(&data[0]), int64(len(data)))

	s.log.Info.Printf("NewShardCUDA(%v) finished\n", s.name)
	return s, nil
}

/*********************************************
 * PUBLIC METHODS
 *********************************************/

// Free releases all OpenCL buffers
func (s *ShardCUDA) Free() error {
	cu.MemFree(s.cudaData)
	s.log.Info.Printf("%v.Free finished\n", s.name)
	return nil
}

// GetBucketSize returns the size (in bytes) of a bucket
func (s *ShardCUDA) GetBucketSize() int {
	return s.bucketSize
}

// GetNumBuckets returns the number of buckets in the shard
func (s *ShardCUDA) GetNumBuckets() int {
	return s.numBuckets
}

// GetData returns a slice of the data
func (s *ShardCUDA) GetData() []byte {
	return s.data[:]
}

// Read handles a batch read, where each request is concatentated into `reqs`
//   each request consists of `reqLength` bytes
//   Note: every request starts on a byte boundary
// Returns: a single byte array where responses are concatenated by the order in `reqs`
//   each response consists of `s.bucketSize` bytes
func (s *ShardCUDA) Read(reqs []byte, reqLength int) ([]byte, error) {
	s.log.Trace.Printf("%v.Read: start\n", s.name)

	if len(reqs)%reqLength != 0 {
		rErr := fmt.Errorf("ShardCUDA.Read expects len(reqs)=%d to be a multiple of reqLength=%d", len(reqs), reqLength)
		s.log.Error.Printf("%v.Read error: %v\n", s.name, rErr)
		return nil, rErr
	}

	inputSize := int64(len(reqs))
	batchSize := inputSize / int64(reqLength)
	outputSize := batchSize * int64(s.bucketSize)
	responses := make([]byte, outputSize)

	// Weird context hack
	if cu.CtxGetCurrent() == 0 {
		s.context.Ctx.SetCurrent()
	}

	// Create buffers
	input := cu.MemAlloc(inputSize)
	defer cu.MemFree(input)
	output := cu.MemAlloc(outputSize)
	defer cu.MemFree(output)

	//free, total := cu.MemGetInfo()
	//s.log.Info.Printf("MemGetInfo: %v / %v bytes free\n", free, total)
	//size, base := cu.MemGetAddressRange(output)
	//s.log.Info.Printf("size=%v, base=%v\n", size, base)
	//s.log.Info.Printf("input=%v, output=%v\n", input.MemoryType().String(), output.MemoryType().String())

	// Copy input to device
	cu.MemcpyHtoD(input, unsafe.Pointer(&reqs[0]), inputSize)
	// Zero output
	//cu.MemcpyHtoD(output, unsafe.Pointer(&responses[0]), outputSize)
	cu.MemsetD8(output, 0, outputSize)

	//Set kernel args
	data := s.cudaData
	batchSize32 := int32(batchSize)
	reqLength32 := int32(reqLength)
	numBuckets32 := int32(s.numBuckets)
	bucketSize32 := int32(s.bucketSize / s.context.GetKernelDataSize())
	local := s.context.GetGroupSize()
	global := s.numThreads
	if global < local {
		local = global
	}
	global32 := int32(global)
	args := []unsafe.Pointer{
		unsafe.Pointer(&data),
		unsafe.Pointer(&input),
		unsafe.Pointer(&output),
		unsafe.Pointer(&batchSize32),
		unsafe.Pointer(&reqLength32),
		unsafe.Pointer(&numBuckets32),
		unsafe.Pointer(&bucketSize32),
		unsafe.Pointer(&global32),
	}

	/** START LOCK REGION **/
	//s.context.KernelMutex.Lock()

	//cu.LaunchKernel(s.context.ZeroFn, (global-1)/local+1, 1, 1, local, 1, 1, 0, 0, args)
	cu.LaunchKernel(s.context.PIRFn, (global-1)/local+1, 1, 1, local, 1, 1, 0, 0, args)
	cu.CtxSynchronize()

	//s.context.KernelMutex.Unlock()
	/** END LOCK REGION **/

	// Read responses
	cu.MemcpyDtoH(unsafe.Pointer(&responses[0]), output, outputSize)

	s.log.Trace.Printf("%v.Read: end \n", s.name)
	return responses, nil
}
