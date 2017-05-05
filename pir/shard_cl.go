//+build !travis

package pir

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/go-gl/cl/v1.2/cl"
	"github.com/privacylab/talek/common"
)

// ShardCL represents a read-only shard of the database,
// backed by an OpenCL implementation of PIR
type ShardCL struct {
	log        *common.Logger
	name       string
	context    *ContextCL
	bucketSize int
	numBuckets int
	data       []byte
	numThreads int
	clData     cl.Mem
}

// NewShardCL creates a new OpenCL-backed shard
// The data is represented as a flat byte array = append(bucket_1, bucket_2 ... bucket_n)
// Pre-conditions:
// - len(data) must be a multiple of bucketSize
// Returns: the shard, or an error if mismatched size
func NewShardCL(name string, context *ContextCL, bucketSize int, data []byte, numThreads int) (*ShardCL, error) {
	s := &ShardCL{}
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

	/** OpenCL **/
	//  Create buffers
	var errptr *cl.ErrorCode
	s.clData = cl.CreateBuffer(s.context.Context, cl.MEM_READ_ONLY, uint64(len(data)), nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		return nil, fmt.Errorf("NewShardCL(%v) failed: couldnt create OpenCL buffer", name)
	}
	//Write shard data to GPU
	err := cl.EnqueueWriteBuffer(s.context.CommandQueue, s.clData, cl.TRUE, 0, uint64(len(data)), unsafe.Pointer(&data[0]), 0, nil, nil)
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("NewShardCL(%v) failed: cannot write shard to GPU (OpenCL buffer)", name)
	}

	return s, nil
}

/*********************************************
 * PUBLIC METHODS
 *********************************************/

// Free releases all OpenCL buffers
func (s *ShardCL) Free() error {
	errStr := ""
	err := cl.ReleaseMemObject(s.clData)
	if err != cl.SUCCESS {
		errStr += cl.ErrToStr(err) + "\n"
	}
	if strings.Compare(errStr, "") != 0 {
		return fmt.Errorf("ContextCL.Free errors: " + errStr)
	}
	return nil
}

// GetName returns the name of the shard
func (s *ShardCL) GetName() string {
	return s.name
}

// GetBucketSize returns the size (in bytes) of a bucket
func (s *ShardCL) GetBucketSize() int {
	return s.bucketSize
}

// GetNumBuckets returns the number of buckets in the shard
func (s *ShardCL) GetNumBuckets() int {
	return s.numBuckets
}

// GetData returns a slice of the data
func (s *ShardCL) GetData() []byte {
	return s.data[:]
}

// Read handles a batch read, where each request is concatentated into `reqs`
//   each request consists of `reqLength` bytes
//   Note: every request starts on a byte boundary
// Returns: a single byte array where responses are concatenated by the order in `reqs`
//   each response consists of `s.bucketSize` bytes
func (s *ShardCL) Read(reqs []byte, reqLength int) ([]byte, error) {
	if len(reqs)%reqLength != 0 {
		return nil, fmt.Errorf("ShardCL.Read expects len(reqs)=%d to be a multiple of reqLength=%d", len(reqs), reqLength)
	}

	inputSize := len(reqs)
	batchSize := inputSize / reqLength
	outputSize := batchSize * s.bucketSize
	responses := make([]byte, outputSize)
	context := s.context.Context
	var err cl.ErrorCode
	var errptr *cl.ErrorCode

	//Create buffers
	input := cl.CreateBuffer(context, cl.MEM_READ_ONLY, uint64(inputSize), nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		return nil, fmt.Errorf("couldnt create input buffer")
	}
	defer cl.ReleaseMemObject(input)

	output := cl.CreateBuffer(context, cl.MEM_WRITE_ONLY, uint64(outputSize), nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		return nil, fmt.Errorf("couldnt create output buffer")
	}
	defer cl.ReleaseMemObject(output)

	//Write request data
	err = cl.EnqueueWriteBuffer(s.context.CommandQueue, input, cl.TRUE, 0, uint64(inputSize), unsafe.Pointer(&reqs[0]), 0, nil, nil)
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to write to input requests (OpenCL buffer)")
	}

	/** START LOCK REGION **/
	s.context.KernelMutex.Lock()

	// Note: SetKernelArgs->EnqueueNDRangeKernel is not thread-safe
	//   @todo - create multiple kernels to support parallel PIR in a single context
	//   https://www.khronos.org/registry/OpenCL/sdk/1.2/docs/man/xhtml/clSetKernelArg.html
	//Set kernel args
	data := s.clData
	err = cl.SetKernelArg(s.context.Kernel, 0, 8, unsafe.Pointer(&data))
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to write kernel arg 0")
	}
	err = cl.SetKernelArg(s.context.Kernel, 1, 8, unsafe.Pointer(&input))
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to write kernel arg 1")
	}
	err = cl.SetKernelArg(s.context.Kernel, 2, 8, unsafe.Pointer(&output))
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to write kernel arg 2")
	}
	err = cl.SetKernelArg(s.context.Kernel, 3, GPU_SCRATCH_SIZE, nil)
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to write kernel arg 3")
	}
	batchSize32 := uint32(batchSize)
	err = cl.SetKernelArg(s.context.Kernel, 4, 4, unsafe.Pointer(&batchSize32))
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to write kernel arg 4")
	}
	reqLength32 := uint32(reqLength)
	err = cl.SetKernelArg(s.context.Kernel, 5, 4, unsafe.Pointer(&reqLength32))
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to write kernel arg 5")
	}
	numBuckets32 := uint32(s.numBuckets)
	err = cl.SetKernelArg(s.context.Kernel, 6, 4, unsafe.Pointer(&numBuckets32))
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to write kernel arg 6")
	}
	bucketSize32 := uint32(s.bucketSize / KERNEL_DATATYPE_SIZE)
	err = cl.SetKernelArg(s.context.Kernel, 7, 4, unsafe.Pointer(&bucketSize32))
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to write kernel arg 7")
	}
	//global := local
	local := uint64(s.context.GetGroupSize())
	global := uint64(s.numThreads)
	if global < local {
		local = global
	}
	global32 := uint32(global)
	err = cl.SetKernelArg(s.context.Kernel, 8, 4, unsafe.Pointer(&global32))
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to write kernel arg 8")
	}
	scratchSize32 := uint32(GPU_SCRATCH_SIZE / KERNEL_DATATYPE_SIZE)
	err = cl.SetKernelArg(s.context.Kernel, 9, 4, unsafe.Pointer(&scratchSize32))
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to write kernel arg 9")
	}

	//s.log.Info.Printf("local=%v, global=%v\n", local, global)
	err = cl.EnqueueNDRangeKernel(s.context.CommandQueue, s.context.Kernel, 1, nil, &global, &local, 0, nil, nil)
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to execute kernel!")
	}
	cl.Finish(s.context.CommandQueue)
	s.context.KernelMutex.Unlock()
	/** END LOCK REGION **/

	err = cl.EnqueueReadBuffer(s.context.CommandQueue, output, cl.TRUE, 0, uint64(outputSize), unsafe.Pointer(&responses[0]), 0, nil, nil)
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to read output response (OpenCL buffer), err=%v", cl.ErrToStr(err))
	}

	return responses, nil
}

/*********************************************
 * PRIVATE METHODS
 *********************************************/
