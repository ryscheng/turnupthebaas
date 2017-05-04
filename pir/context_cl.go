//+build !travis

package pir

import (
	"fmt"
	"io/ioutil"
	"strings"
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
	kernelSource string
	platformID   cl.PlatformID
	deviceID     cl.DeviceId
	Context      cl.Context
	CommandQueue cl.CommandQueue
	program      cl.Program
	Kernel       cl.Kernel
	groupSize    uint64
}

func NewContextCL(name string, kernelFile string) (*ContextCL, error) {
	c := &ContextCL{}
	c.log = common.NewLogger(name)
	c.name = name

	// Read Kernel Source
	kernelBytes, fErr := ioutil.ReadFile(kernelFile)
	if fErr != nil {
		return nil, fmt.Errorf("NewContextCl: failed to read kernel source")
	}
	c.kernelSource = string(kernelBytes) + "\x00"

	// Get Platform
	ids := make([]cl.PlatformID, 100)
	count := uint32(0)
	err := cl.GetPlatformIDs(uint32(len(ids)), &ids[0], &count)
	if err != cl.SUCCESS || count < 1 {
		return nil, fmt.Errorf("NewContextCl: failed to retrieve OpenCL platform ID\n")
	}
	c.platformID = ids[0]

	// Get Device
	var device cl.DeviceId
	err = cl.GetDeviceIDs(c.platformID, cl.DEVICE_TYPE_GPU, 1, &device, &count)
	if err != cl.SUCCESS || count < 1 {
		return nil, fmt.Errorf("NewContextCl: failed to create OpenCL device group\n")
	}
	c.deviceID = device

	//Create Computer Context
	var errptr *cl.ErrorCode
	c.Context = cl.CreateContext(nil, 1, &device, nil, nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		c.log.Error.Fatal("couldnt create context")
	}

	//Create Command Queue
	c.CommandQueue = cl.CreateCommandQueue(c.Context, device, 0, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		c.log.Error.Fatal("couldnt create command queue")
	}

	//Create program
	srcptr := cl.Str(c.kernelSource)
	c.program = cl.CreateProgramWithSource(c.Context, 1, &srcptr, nil, errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		c.log.Error.Fatal("couldnt create program")
	}

	err = cl.BuildProgram(c.program, 1, &device, nil, nil, nil)
	if err != cl.SUCCESS {
		var length uint64
		buffer := make([]byte, 1024)
		c.log.Error.Println("Error: Failed to build program executable!")
		cl.GetProgramBuildInfo(c.program, device, cl.PROGRAM_BUILD_LOG, uint64(len(buffer)), unsafe.Pointer(&buffer[0]), &length)
		c.log.Error.Fatal(string(buffer[0:length]))
	}

	//Get Kernel (~CUDA Grid)
	c.Kernel = cl.CreateKernel(c.program, cl.Str("pir"+"\x00"), errptr)
	if errptr != nil && cl.ErrorCode(*errptr) != cl.SUCCESS {
		c.log.Error.Fatal("couldnt create compute kernel")
	}

	// OpenCL work-group = CUDA block
	groupSize := uint64(0)
	err = cl.GetKernelWorkGroupInfo(c.Kernel, device, cl.KERNEL_WORK_GROUP_SIZE, 8, unsafe.Pointer(&groupSize), nil)
	if err != cl.SUCCESS {
		c.log.Error.Fatal("Failed to get kernel work group info")
	}
	c.groupSize = groupSize

	return c, nil

}

/*********************************************
 * PUBLIC METHODS
 *********************************************/

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
		return fmt.Errorf("ContextCL.Free errors: " + errStr)
	}
	return nil
}

// Returns the working group size of this context
func (c *ContextCL) GetGroupSize() uint64 {
	return c.groupSize
}

/*********************************************
 * PRIVATE METHODS
 *********************************************/
