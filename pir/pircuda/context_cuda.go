//+build !nocuda,!travis

package pircuda

import (
	"sync"

	"github.com/barnex/cuda5/cu"
	"github.com/privacylab/talek/common"
)

const (
	// CudaDeviceID is the hardcoded GPU device we'll use
	CudaDeviceID = 0
)

// ContextCUDA represents a single CUDA context
// Currently, we only support 1 live ShardCUDA per ContextCUDA
// Before creating a new ShardCUDA, the old one must be Free()
type ContextCUDA struct {
	log            *common.Logger
	name           string
	KernelMutex    *sync.Mutex
	kernelSource   string
	kernelDataSize int
	device         cu.Device
	Ctx            cu.Context
	module         cu.Module
	PIRFn          cu.Function
	groupSize      int
}

// NewContextCUDA creates a new CUDA context, shared among ShardCUDA instances
func NewContextCUDA(name string, kernelSource string, kernelDataSize int) (*ContextCUDA, error) {
	c := &ContextCUDA{}
	c.log = common.NewLogger(name)
	c.name = name
	c.KernelMutex = &sync.Mutex{}
	c.kernelSource = kernelSource
	c.kernelDataSize = kernelDataSize

	cu.Init(0)
	c.device = cu.DeviceGet(CudaDeviceID)
	c.groupSize = c.device.Attribute(cu.MAX_THREADS_PER_BLOCK)
	c.Ctx = cu.CtxCreate(cu.CTX_SCHED_AUTO, c.device)
	c.Ctx.SetCurrent()

	//major, minor := c.device.ComputeCapability()
	//c.log.Info.Printf("CUDA Compute Compatibility: %v.%v\n", major, minor)

	//c.module = cu.ModuleLoadData(kernelSource)
	c.module = cu.ModuleLoad(kernelSource)
	c.PIRFn = c.module.GetFunction("pir")

	// Weird hack in context handling
	if cu.CtxGetCurrent() == 0 && c.Ctx != 0 {
		c.Ctx.SetCurrent()
	}

	return c, nil
}

/*********************************************
 * PUBLIC METHODS
 *********************************************/

// GetName returns the name of the context
func (c *ContextCUDA) GetName() string {
	return c.name
}

// GetGroupSize returns the working group size of this context
func (c *ContextCUDA) GetGroupSize() int {
	return c.groupSize
}

// GetKernelDataSize returns the size (in bytes) of a single data item in the kernel
func (c *ContextCUDA) GetKernelDataSize() int {
	return c.kernelDataSize
}

// Free currently does nothing. ShardCL waits for the go garbage collector
func (c *ContextCUDA) Free() error {
	c.Ctx.Destroy()
	return nil
}
