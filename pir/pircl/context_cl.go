//+build opencl,!travis

package pircl

import (
	"fmt"
	"strings"
	"sync"
	"unsafe"

	"github.com/go-gl/cl/v1.2/cl"
	"github.com/privacylab/talek/common"
)

// ContextCL represents a single OpenCL context
// Currently, we only support 1 live ShardCL per ContextCL
// Before creating a new ShardCL, the old one must be Free()
type ContextCL struct {
	log            *common.Logger
	name           string
	kernelSource   string
	kernelDataSize int
	gpuScratchSize int
	KernelMutex    *sync.Mutex
	platformID     cl.PlatformID
	deviceID       cl.DeviceId
	Context        cl.Context
	CommandQueue   cl.CommandQueue
	program        cl.Program
	Kernel         cl.Kernel
	groupSize      int
}

// NewContextCL creates a new OpenCL context with a given kernel source.
// New ShardCL instances will share the same kernel
func NewContextCL(name string, kernelSource string, kernelDataSize int, gpuScratchSize int) (*ContextCL, error) {
	c := &ContextCL{}
	c.log = common.NewLogger(name)
	c.name = name
	c.kernelSource = kernelSource
	c.kernelDataSize = kernelDataSize
	c.gpuScratchSize = gpuScratchSize
	c.KernelMutex = &sync.Mutex{}

	// Get Platform
	ids := make([]cl.PlatformID, 100)
	count := uint32(0)
	err := cl.GetPlatformIDs(uint32(len(ids)), &ids[0], &count)
	if err != cl.SUCCESS || count < 1 {
		c.Free()
		rErr := fmt.Errorf("NewContextCl: failed to retrieve OpenCL platform ID")
		c.log.Error.Printf("NewContextCL(%v) error: %v\n", c.name, rErr)
		return nil, rErr
	}
	c.platformID = ids[0]

	// Get Device
	var device cl.DeviceId
	err = cl.GetDeviceIDs(c.platformID, cl.DEVICE_TYPE_GPU, 1, &device, &count)
	if err != cl.SUCCESS || count < 1 {
		c.Free()
		rErr := fmt.Errorf("NewContextCl: failed to create OpenCL device group")
		c.log.Error.Printf("NewContextCL(%v) error: %v\n", c.name, rErr)
		return nil, rErr
	}
	c.deviceID = device

	//Create Computer Context
	var errptr *cl.ErrorCode
	notify := func(arg1 string, arg2 unsafe.Pointer, arg3 uint64, arg4 interface{}) {
		fmt.Printf("OpenCL Context Error: %v %v %v %v \n", arg1, arg2, arg3, arg4)
	}
	c.Context = cl.CreateContext(nil, 1, &device, notify, nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		c.Free()
		rErr := fmt.Errorf("NewContextCl: couldnt create context")
		c.log.Error.Printf("NewContextCL(%v) error: %v\n", c.name, rErr)
		return nil, rErr
	}

	//Create Command Queue
	c.CommandQueue = cl.CreateCommandQueue(c.Context, device, 0, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		c.Free()
		rErr := fmt.Errorf("NewContextCl: couldnt create command queue")
		c.log.Error.Printf("NewContextCL(%v) error: %v\n", c.name, rErr)
		return nil, rErr
	}

	//Create program
	//srcptr := cl.Str("__kernel\nvoid pir() {}" + "\x00")
	srcptr := cl.Str(c.kernelSource)
	// Read Kernel Source
	c.program = cl.CreateProgramWithSource(c.Context, 1, &srcptr, nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		c.Free()
		rErr := fmt.Errorf("NewContextCl: couldnt create program")
		c.log.Error.Printf("NewContextCL(%v) error: %v\n", c.name, rErr)
		return nil, rErr
	}

	err = cl.BuildProgram(c.program, 1, &device, nil, nil, nil)
	if err != cl.SUCCESS {
		var length uint64
		buffer := make([]byte, 2048)
		c.log.Error.Println("NewContextCl Error: Failed to build program executable!")
		cl.GetProgramBuildInfo(c.program, device, cl.PROGRAM_BUILD_LOG, uint64(len(buffer)), unsafe.Pointer(&buffer[0]), &length)
		c.Free()
		var rErr error
		if length < uint64(len(buffer)) {
			rErr = fmt.Errorf(string(buffer[0:length]))
		} else {
			rErr = fmt.Errorf(string(buffer[0:]))
		}
		c.log.Error.Printf("NewContextCL(%v) error: %v\n", c.name, rErr)
		return nil, rErr
	}

	//Get Kernel (~CUDA Grid)
	c.Kernel = cl.CreateKernel(c.program, cl.Str("pir"+"\x00"), errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		c.Free()
		rErr := fmt.Errorf("NewContextCl: couldnt create compute kernel")
		c.log.Error.Printf("NewContextCL(%v) error: %v\n", c.name, rErr)
		return nil, rErr
	}

	// OpenCL work-group = CUDA block
	groupSize := uint64(0)
	err = cl.GetKernelWorkGroupInfo(c.Kernel, device, cl.KERNEL_WORK_GROUP_SIZE, 8, unsafe.Pointer(&groupSize), nil)
	if err != cl.SUCCESS {
		c.Free()
		rErr := fmt.Errorf("NewContextCl: Failed to get kernel work group info")
		c.log.Error.Printf("NewContextCL(%v) error: %v\n", c.name, rErr)
		return nil, rErr
	}
	c.groupSize = int(groupSize)

	c.log.Info.Printf("NewContextCL(%v) finished\n", c.name)
	return c, nil

}

/*********************************************
 * PUBLIC METHODS
 *********************************************/

// GetGroupSize returns the working group size of this context
func (c *ContextCL) GetGroupSize() int {
	return c.groupSize
}

// GetKernelDataSize returns the size (in bytes) of a single data item in the kernel
func (c *ContextCL) GetKernelDataSize() int {
	return c.kernelDataSize
}

// Free currently does nothing. ShardCL waits for the go garbage collector
func (c *ContextCL) Free() error {
	errStr := ""
	err := cl.ReleaseKernel(c.Kernel)
	if err != cl.SUCCESS {
		errStr += cl.ErrToStr(err) + "\n"
	}
	err = cl.ReleaseProgram(c.program)
	if err != cl.SUCCESS {
		errStr += cl.ErrToStr(err) + "\n"
	}
	cl.ReleaseCommandQueue(c.CommandQueue)
	err = cl.ReleaseContext(c.Context)
	if err != cl.SUCCESS {
		errStr += cl.ErrToStr(err) + "\n"
	}
	if strings.Compare(errStr, "") != 0 {
		c.log.Error.Printf("%v.Free error: %v\n", c.name, errStr)
		return fmt.Errorf("ContextCL.Free errors: " + errStr)
	}
	c.log.Info.Printf("%v.Free finished\n", c.name)
	return nil
}

// GetGPUScratchSize returns the size of the scratch used by the kernel
func (c *ContextCL) GetGPUScratchSize() int {
	return c.gpuScratchSize
}
