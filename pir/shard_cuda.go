//+build !travis

package pir

import (
	"fmt"
	"unsafe"

	"github.com/barnex/cuda5/cu"
	"github.com/privacylab/talek/common"
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
		return nil, fmt.Errorf("NewShardCL(%v) failed: data(len=%v) not multiple of bucketSize=%v", name, len(data), bucketSize)
	}

	s.bucketSize = bucketSize
	s.numBuckets = (len(data) / bucketSize)
	s.data = data
	s.numThreads = numThreads

	/** CUDA **/
	//  Create buffers
	s.cudaData = cu.MemAlloc(int64(len(data)))
	cu.MemcpyHtoD(s.cudaData, unsafe.Pointer(&data[0]), int64(len(data)))

	return s, nil
}

/*********************************************
 * PUBLIC METHODS
 *********************************************/

// Free releases all OpenCL buffers
func (s *ShardCUDA) Free() error {
	cu.MemFree(s.cudaData)
	return nil
}

// GetName returns the name of the shard
func (s *ShardCUDA) GetName() string {
	return s.name
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
	if len(reqs)%reqLength != 0 {
		return nil, fmt.Errorf("ShardCUDA.Read expects len(reqs)=%d to be a multiple of reqLength=%d", len(reqs), reqLength)
	}

	inputSize := int64(len(reqs))
	batchSize := inputSize / int64(reqLength)
	outputSize := batchSize * int64(s.bucketSize)
	responses := make([]byte, outputSize)

	// Create buffers
	input := cu.MemAlloc(inputSize)
	defer cu.MemFree(input)
	output := cu.MemAlloc(outputSize)
	defer cu.MemFree(output)

	// Copy input to device
	cu.MemcpyHtoD(input, unsafe.Pointer(&reqs[0]), inputSize)

	//Set kernel args
	data := s.cudaData
	batchSize32 := uint32(batchSize)
	reqLength32 := uint32(reqLength)
	numBuckets32 := uint32(s.numBuckets)
	bucketSize32 := uint32(s.bucketSize / KernelDataSize)
	//global := local
	local := s.context.GetGroupSize()
	global := s.numThreads
	if global < local {
		local = global
	}
	global32 := uint32(global)
	//scratchSize32 := uint32(GPUScratchSize / KernelDataSize)
	//argSizes := []uint64{8, 8, 8, GPUScratchSize, 4, 4, 4, 4, 4, 4}
	args := []unsafe.Pointer{
		unsafe.Pointer(&data),
		unsafe.Pointer(&input),
		unsafe.Pointer(&output),
		//nil,
		unsafe.Pointer(&batchSize32),
		unsafe.Pointer(&reqLength32),
		unsafe.Pointer(&numBuckets32),
		unsafe.Pointer(&bucketSize32),
		unsafe.Pointer(&global32),
		//unsafe.Pointer(&scratchSize32),
	}

	/** START LOCK REGION **/
	s.context.KernelMutex.Lock()

	cu.LaunchKernel(s.context.Fn, global, 1, 1, local, 1, 1, 0, 0, args)
	cu.CtxSynchronize()

	s.context.KernelMutex.Unlock()
	/** END LOCK REGION **/

	// Read responses
	cu.MemcpyDtoH(unsafe.Pointer(&responses[0]), output, outputSize)

	return responses, nil
}

/*********************************************
 * PRIVATE METHODS
 *********************************************/
