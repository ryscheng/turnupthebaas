//+build !travis

package pir

import (
	"fmt"
	//"math"
	"strings"
	"unsafe"

	"github.com/go-gl/cl/v1.2/cl"
	"github.com/privacylab/talek/common"
)

// ShardCL represents a read-only shard of the database,
// backed by an OpenCL implementation of PIR
type ShardCL struct {
	log         *common.Logger
	name        string
	context     *ContextCL
	bucketSize  int
	numBuckets  int
	data        []byte
	readVersion int
	clData      cl.Mem
}

// NewShardCL creates a new OpenCL-backed shard
// The data is represented as a flat byte array = append(bucket_1, bucket_2 ... bucket_n)
// Pre-conditions:
// - len(data) must be a multiple of bucketSize
// Returns: the shard, or an error if mismatched size
func NewShardCL(name string, context *ContextCL, bucketSize int, data []byte, readVersion int) (*ShardCL, error) {
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
	s.readVersion = readVersion

	/** OpenCL **/
	//  Create buffers
	var errptr *cl.ErrorCode
	s.clData = cl.CreateBuffer(s.context.Context, cl.MEM_READ_ONLY, uint64(len(data)), nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		return nil, fmt.Errorf("NewShardCL(%v) failed: couldnt create OpenCL buffer", name)
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
	} else if s.readVersion == 0 {
		return s.read0(reqs, reqLength)
	}

	return nil, fmt.Errorf("ShardCL.Read: invalid readVersion=%d", s.readVersion)
}

/*********************************************
 * PRIVATE METHODS
 *********************************************/

func (s *ShardCL) read0(reqs []byte, reqLength int) ([]byte, error) {
	inputSize := len(reqs)
	numReqs := inputSize / reqLength
	outputSize := numReqs * s.bucketSize
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

	//Set kernel args
	count := uint32(DataSize)
	err = cl.SetKernelArg(s.context.Kernel, 0, 8, unsafe.Pointer(&input))
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to write kernel arg 0")
	}
	err = cl.SetKernelArg(s.context.Kernel, 1, 8, unsafe.Pointer(&output))
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to write kernel arg 1")
	}
	err = cl.SetKernelArg(s.context.Kernel, 2, 4, unsafe.Pointer(&count))
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to write kernel arg 2")
	}

	//global := local
	local := uint64(s.context.GetGroupSize())
	global := uint64(DataSize)
	s.log.Info.Printf("local=%v, global=%v\n", local, global)
	err = cl.EnqueueNDRangeKernel(s.context.CommandQueue, s.context.Kernel, 1, nil, &global, &local, 0, nil, nil)
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to execute kernel!")
	}
	cl.Finish(s.context.CommandQueue)

	err = cl.EnqueueReadBuffer(s.context.CommandQueue, output, cl.TRUE, 0, uint64(outputSize), unsafe.Pointer(&responses[0]), 0, nil, nil)
	if err != cl.SUCCESS {
		return nil, fmt.Errorf("Failed to read output response (OpenCL buffer)!")
	}

	return responses, nil
}
