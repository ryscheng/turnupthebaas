//+build !noopencl,!travis

package pircl

import (
	"fmt"
	"strconv"
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

// NewShard creates a new cuda shard conforming to the common interface
func NewShard(bucketSize int, data []byte, userdata string) pir.Shard {
	parts := strings.Split(userdata, ".")
	if len(parts) < 6 {
		fmt.Errorf("Invalid cl specification: %s. Should be cl.[source].[datasize].[scratchsize].[threads]", parts)
		return nil
	}

	dataSize, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil {
		fmt.Errorf("Invalid datasize: %s. Should be numeric", parts[2])
		return nil
	}
	scratchSize, err := strconv.ParseInt(parts[3], 10, 32)
	if err != nil {
		fmt.Errorf("Invalid scratch size: %s. Should be numeric", parts[3])
		return nil
	}

	source := KernelCL0
	if parts[1] == "1" {
		source = KernelCL1
	} else if parts[1] == "2" {
		source = KernelCL2
	}

	context, err := NewContextCL("contextcl", source, int(dataSize), int(scratchSize))
	if err != nil {
		fmt.Fatalf("cannot create new ContextCL: error=%v\n", err)
	}

	threads, err := strconv.ParseInt(parts[4], 10, 32)
	if err != nil {
		fmt.Errorf("Invalid threads: %s. Should be numeric", parts[4])
		return nil
	}
	shard, err := NewShardCL("CL Shard ("+userdata+")", context, bucketSize, data, int(threads))
	if err != nil {
		fmt.Errorf("Could not create CL shard: %v", err)
		return nil
	}
	return pir.Shard(shard)
}

func init() {
	pir.PIRBackings["cl"] = NewShard
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
		rErr := fmt.Errorf("NewShardCL(%v) failed: data(len=%v) not multiple of bucketSize=%v", name, len(data), bucketSize)
		s.log.Error.Printf("%v\n", rErr)
		return nil, rErr
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
		rErr := fmt.Errorf("NewShardCL(%v) failed: couldnt create OpenCL buffer", name)
		s.log.Error.Printf("%v\n", rErr)
		return nil, rErr
	}
	//Write shard data to GPU
	err := cl.EnqueueWriteBuffer(s.context.CommandQueue, s.clData, cl.TRUE, 0, uint64(len(data)), unsafe.Pointer(&data[0]), 0, nil, nil)
	if err != cl.SUCCESS {
		rErr := fmt.Errorf("NewShardCL(%v) failed: cannot write shard to GPU (OpenCL buffer)", name)
		s.log.Error.Printf("%v\n", rErr)
		return nil, rErr
	}

	s.log.Info.Printf("NewShardCL(%v) finished\n", s.name)
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
		s.log.Error.Printf("%v.Free error: %v\n", s.name, errStr)
		return fmt.Errorf("ContextCL.Free errors: " + errStr)
	}
	s.log.Info.Printf("%v.Free finished\n", s.name)
	return nil
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
	s.log.Trace.Printf("%v.Read: start\n", s.name)
	if len(reqs)%reqLength != 0 {
		rErr := fmt.Errorf("ShardCL.Read expects len(reqs)=%d to be a multiple of reqLength=%d", len(reqs), reqLength)
		s.log.Error.Printf("%v.Read error: %v\n", s.name, rErr)
		return nil, rErr
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
		rErr := fmt.Errorf("couldnt create input buffer")
		s.log.Error.Printf("%v.Read error: %v\n", s.name, rErr)
		return nil, rErr
	}
	defer cl.ReleaseMemObject(input)

	output := cl.CreateBuffer(context, cl.MEM_WRITE_ONLY, uint64(outputSize), nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		rErr := fmt.Errorf("couldnt create output buffer")
		s.log.Error.Printf("%v.Read error: %v\n", s.name, rErr)
		return nil, rErr
	}
	defer cl.ReleaseMemObject(output)

	//Write request data
	err = cl.EnqueueWriteBuffer(s.context.CommandQueue, input, cl.TRUE, 0, uint64(inputSize), unsafe.Pointer(&reqs[0]), 0, nil, nil)
	if err != cl.SUCCESS {
		rErr := fmt.Errorf("Failed to write to input requests (OpenCL buffer)")
		s.log.Error.Printf("%v.Read error: %v\n", s.name, rErr)
		return nil, rErr
	}

	//Set kernel args
	data := s.clData
	batchSize32 := uint32(batchSize)
	reqLength32 := uint32(reqLength)
	numBuckets32 := uint32(s.numBuckets)
	bucketSize32 := uint32(s.bucketSize / s.context.GetKernelDataSize())
	//global := local
	local := uint64(s.context.GetGroupSize())
	global := uint64(s.numThreads)
	if global < local {
		local = global
	}
	global32 := uint32(global)
	scratchSize32 := uint32(s.context.GetGPUScratchSize() / s.context.GetKernelDataSize())
	argSizes := []uint64{8, 8, 8, uint64(s.context.GetGPUScratchSize()), 4, 4, 4, 4, 4, 4}
	args := []unsafe.Pointer{
		unsafe.Pointer(&data),
		unsafe.Pointer(&input),
		unsafe.Pointer(&output),
		nil,
		unsafe.Pointer(&batchSize32),
		unsafe.Pointer(&reqLength32),
		unsafe.Pointer(&numBuckets32),
		unsafe.Pointer(&bucketSize32),
		unsafe.Pointer(&global32),
		unsafe.Pointer(&scratchSize32),
	}

	/** START LOCK REGION **/
	s.context.KernelMutex.Lock()
	// Note: SetKernelArgs->EnqueueNDRangeKernel is not thread-safe
	//   @todo - create multiple kernels to support parallel PIR in a single context
	//   https://www.khronos.org/registry/OpenCL/sdk/1.2/docs/man/xhtml/clSetKernelArg.html

	for i := 0; i < len(args); i++ {
		err = cl.SetKernelArg(s.context.Kernel, uint32(i), argSizes[i], args[i])
		if err != cl.SUCCESS {
			rErr := fmt.Errorf("Failed to write kernel arg %v", i)
			s.log.Error.Printf("%v.Read error: %v\n", s.name, rErr)
			return nil, rErr
		}
	}

	//s.log.Info.Printf("local=%v, global=%v\n", local, global)
	err = cl.EnqueueNDRangeKernel(s.context.CommandQueue, s.context.Kernel, 1, nil, &global, &local, 0, nil, nil)
	if err != cl.SUCCESS {
		rErr := fmt.Errorf("Failed to execute kernel")
		s.log.Error.Printf("%v.Read error: %v\n", s.name, rErr)
		return nil, rErr
	}
	s.context.KernelMutex.Unlock()
	/** END LOCK REGION **/

	cl.Finish(s.context.CommandQueue) //@todo inside or outside lock region?

	err = cl.EnqueueReadBuffer(s.context.CommandQueue, output, cl.TRUE, 0, uint64(outputSize), unsafe.Pointer(&responses[0]), 0, nil, nil)
	if err != cl.SUCCESS {
		rErr := fmt.Errorf("Failed to read output response (OpenCL buffer), err=%v", cl.ErrToStr(err))
		s.log.Error.Printf("%v.Read error: %v\n", s.name, rErr)
		return nil, rErr
	}

	s.log.Trace.Printf("%v.Read: end\n", s.name)
	return responses, nil
}
