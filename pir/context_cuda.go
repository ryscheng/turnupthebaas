//+build !travis

package pir

import (
	//"fmt"
	//"strings"
	"sync"
	//"unsafe"

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
	log          *common.Logger
	name         string
	KernelMutex  *sync.Mutex
	kernelSource string
	device       cu.Device
	ctx          cu.Context
	module       cu.Module
	Fn           cu.Function
	groupSize    int
}

// NewContextCUDA creates a new CUDA context, shared among ShardCUDA instances
func NewContextCUDA(name string, kernelSource string) (*ContextCUDA, error) {
	c := &ContextCUDA{}
	c.log = common.NewLogger(name)
	c.name = name
	c.KernelMutex = &sync.Mutex{}
	c.kernelSource = kernelSource

	cu.Init(0)
	c.device = cu.DeviceGet(CudaDeviceID)
	c.groupSize = c.device.Attribute(cu.MAX_THREADS_PER_BLOCK)
	c.ctx = cu.CtxCreate(cu.CTX_SCHED_AUTO, c.device)
	c.ctx.SetCurrent()

	//c.module = cu.ModuleLoadData(kernelSource)
	c.module = cu.ModuleLoad(kernelSource)
	c.Fn = c.module.GetFunction("pir")

	// Weird hack in context handling
	if cu.CtxGetCurrent() == 0 && c.ctx != 0 {
		c.ctx.SetCurrent()
	}

	return c, nil
}

/*********************************************
 * PUBLIC METHODS
 *********************************************/

// Free currently does nothing. ShardCL waits for the go garbage collector
func (c *ContextCUDA) Free() error {
	c.ctx.Destroy()
	return nil
}

// GetGroupSize returns the working group size of this context
func (c *ContextCUDA) GetGroupSize() int {
	return c.groupSize
}

/*********************************************
 * PRIVATE METHODS
 *********************************************/
