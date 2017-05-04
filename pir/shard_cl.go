//+build !travis

package pir

import (
	"fmt"
	//"math"
	//"math/rand"
	//"unsafe"

	//"github.com/go-gl/cl/v1.2/cl"
	"github.com/privacylab/talek/common"
)

// ShardCL represents a read-only shard of the database,
// backed by an OpenCL implementation of PIR
type ShardCL struct {
	log         *common.Logger
	name        string
	bucketSize  int
	numBuckets  int
	data        []byte
	readVersion int
	context     *ContextCL
}

const (
	// DataSize is the size of the data we're going to pass to the CL device.
	DataSize = 1024
)

// NewShardCL creates a new OpenCL-backed shard
// The data is represented as a flat byte array = append(bucket_1, bucket_2 ... bucket_n)
// Pre-conditions:
// - len(data) must be a multiple of bucketSize
// Returns: the shard, or an error if mismatched size
func NewShardCL(name string, bucketSize int, data []byte, readVersion int, context *ContextCL) (*ShardCL, error) {
	s := &ShardCL{}
	s.log = common.NewLogger(name)
	s.name = name

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
	s.context = context

	return s, nil
}

/*********************************************
 * PUBLIC METHODS
 *********************************************/

// Free currently does nothing. ShardCL waits for the go garbage collector
func (s *ShardCL) Free() error {
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

func (s *ShardCL) read0(reqs []byte, reqLength int) ([]byte, error) {
	numReqs := len(reqs) / reqLength
	responses := make([]byte, numReqs*s.bucketSize)
	/**
		device := s.deviceId

		// calculate PIR
		data := make([]float32, DataSize)
		for x := 0; x < len(data); x++ {
			data[x] = rand.Float32()*99 + 1
		}

		//Create Computer Context
		var errptr *cl.ErrorCode
		context := cl.CreateContext(nil, 1, &device, nil, nil, errptr)
		if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
			s.log.Error.Fatal("couldnt create context")
		}
		defer cl.ReleaseContext(context)

		//Create Command Queue
		cq := cl.CreateCommandQueue(context, device, 0, errptr)
		if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
			s.log.Error.Fatal("couldnt create command queue")
		}
		defer cl.ReleaseCommandQueue(cq)

		//Create program
		srcptr := cl.Str(KernelSource)
		program := cl.CreateProgramWithSource(context, 1, &srcptr, nil, errptr)
		if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
			s.log.Error.Fatal("couldnt create program")
		}
		defer cl.ReleaseProgram(program)

		err := cl.BuildProgram(program, 1, &device, nil, nil, nil)
		if err != cl.SUCCESS {
			var length uint64
			buffer := make([]byte, DataSize)

			s.log.Error.Println("Error: Failed to build program executable!")
			cl.GetProgramBuildInfo(program, device, cl.PROGRAM_BUILD_LOG, uint64(len(buffer)), unsafe.Pointer(&buffer[0]), &length)
			s.log.Error.Fatal(string(buffer[0:length]))
		}

		//Get Kernel (~CUDA Grid)
		kernel := cl.CreateKernel(program, cl.Str("pir"+"\x00"), errptr)
		if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
			s.log.Error.Fatal("couldnt create compute kernel")
		}
		defer cl.ReleaseKernel(kernel)

		//Create buffers
		input := cl.CreateBuffer(context, cl.MEM_READ_ONLY, 4*DataSize, nil, errptr)
		if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
			s.log.Error.Fatal("couldnt create input buffer")
		}
		defer cl.ReleaseMemObject(input)

		output := cl.CreateBuffer(context, cl.MEM_WRITE_ONLY, 4*DataSize, nil, errptr)
		if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
			s.log.Error.Fatal("couldnt create output buffer")
		}
		defer cl.ReleaseMemObject(output)

		//Write data
		err = cl.EnqueueWriteBuffer(cq, input, cl.TRUE, 0, 4*DataSize, unsafe.Pointer(&data[0]), 0, nil, nil)
		if err != cl.SUCCESS {
			s.log.Error.Fatal("Failed to write to source array")
		}

		//Set kernel args
		count := uint32(DataSize)
		err = cl.SetKernelArg(kernel, 0, 8, unsafe.Pointer(&input))
		if err != cl.SUCCESS {
			s.log.Error.Fatal("Failed to write kernel arg 0")
		}
		err = cl.SetKernelArg(kernel, 1, 8, unsafe.Pointer(&output))
		if err != cl.SUCCESS {
			s.log.Error.Fatal("Failed to write kernel arg 1")
		}
		err = cl.SetKernelArg(kernel, 2, 4, unsafe.Pointer(&count))
		if err != cl.SUCCESS {
			s.log.Error.Fatal("Failed to write kernel arg 2")
		}

		// OpenCL work-group = CUDA block
		local := uint64(0)
		err = cl.GetKernelWorkGroupInfo(kernel, device, cl.KERNEL_WORK_GROUP_SIZE, 8, unsafe.Pointer(&local), nil)
		if err != cl.SUCCESS {
			s.log.Error.Fatal("Failed to get kernel work group info")
		}

		//global := local
		global := uint64(DataSize)
		s.log.Info.Printf("local=%v, global=%v\n", local, global)
		err = cl.EnqueueNDRangeKernel(cq, kernel, 1, nil, &global, &local, 0, nil, nil)
		if err != cl.SUCCESS {
			s.log.Error.Fatal("Failed to execute kernel!")
		}

		cl.Finish(cq)

		results := make([]float32, DataSize)
		err = cl.EnqueueReadBuffer(cq, output, cl.TRUE, 0, 4*1024, unsafe.Pointer(&results[0]), 0, nil, nil)
		if err != cl.SUCCESS {
			s.log.Error.Fatal("Failed to read buffer!")
		}

		success := 0
		notzero := 0
		for i, x := range data {
			if math.Abs(float64(x*x-results[i])) < 0.5 {
				success++
			}
			if results[i] > 0 {
				notzero++
			}
			//s.log.Info.Printf("I/O: %f\t%f", x, results[i])
		}

		s.log.Info.Printf("%d/%d success", success, DataSize)
		s.log.Info.Printf("values not zero: %d", notzero)

	**/
	return responses, nil
}
