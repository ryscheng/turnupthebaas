//+build !travis

package pir

import (
	"fmt"
	"unsafe"

	"github.com/go-gl/cl/v1.2/cl"
	"github.com/privacylab/talek/common"
)

// ContextCL represents a single OpenCL context
// Currently, we only support 1 live ShardCL per ContextCL
// Before creating a new ShardCL, the old one must be Free()
type ContextCL struct {
	log          *common.Logger
	name         string
	platformID   cl.PlatformID
	deviceID     cl.DeviceId
	context      cl.Context
	commandQueue cl.CommandQueue
	program      cl.Program
	kernel       cl.Kernel
}

func NewContextCL(name string) (*ContextCL, error) {
	c := &ContextCL{}
	c.log = common.NewLogger(name)
	c.name = name

	// Get Platform
	ids := make([]cl.PlatformID, 100)
	count := uint32(0)
	err := cl.GetPlatformIDs(uint32(len(ids)), &ids[0], &count)
	if err != cl.SUCCESS || count < 1 {
		return nil, fmt.Errorf("NewContextCl: failed to retrieve OpenCL platform ID\n")
	}
	c.platformID = ids[0]

	// Get Device
	devices := make([]cl.DeviceId, 100)
	err = cl.GetDeviceIDs(c.platformID, cl.DEVICE_TYPE_GPU, 1, &devices[0], &count)
	if err != cl.SUCCESS || count < 1 {
		return nil, fmt.Errorf("NewContextCl: failed to create OpenCL device group\n")
	}
	c.deviceID = devices[0]

	//Create Computer Context
	var errptr *cl.ErrorCode
	c.context = cl.CreateContext(nil, 1, &c.deviceID, nil, nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		c.log.Error.Fatal("couldnt create context")
	}

	//Create Command Queue
	c.commandQueue = cl.CreateCommandQueue(c.context, c.deviceID, 0, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		c.log.Error.Fatal("couldnt create command queue")
	}

	//Create program
	srcptr := cl.Str(KernelSource)
	c.program = cl.CreateProgramWithSource(c.context, 1, &srcptr, nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		c.log.Error.Fatal("couldnt create program")
	}

	err = cl.BuildProgram(c.program, 1, &c.deviceID, nil, nil, nil)
	if err != cl.SUCCESS {
		var length uint64
		buffer := make([]byte, DataSize)
		c.log.Error.Println("Error: Failed to build program executable!")
		cl.GetProgramBuildInfo(c.program, c.deviceID, cl.PROGRAM_BUILD_LOG, uint64(len(buffer)), unsafe.Pointer(&buffer[0]), &length)
		c.log.Error.Fatal(string(buffer[0:length]))
	}

	//Get Kernel (~CUDA Grid)
	c.kernel = cl.CreateKernel(c.program, cl.Str("pir"+"\x00"), errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		c.log.Error.Fatal("couldnt create compute kernel")
	}

	return c, nil

}

/*********************************************
 * PUBLIC METHODS
 *********************************************/

// Free currently does nothing. ShardCL waits for the go garbage collector
func (c *ContextCL) Free() error {
	cl.ReleaseKernel(c.kernel)
	cl.ReleaseProgram(c.program)
	cl.ReleaseCommandQueue(c.commandQueue)
	cl.ReleaseContext(c.context)
	return nil
}

/*********************************************
 * PRIVATE METHODS
 *********************************************/
