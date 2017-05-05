package pir

import (
	"fmt"
	"strings"
	"sync"
	"unsafe"

	"github.com/barnex/cuda5/cu"
	"github.com/privacylab/talek/common"
)

const (
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
	fn           cu.Function
	groupSize    uint64
}

func NewContextCL(name string, kernelSource string) (*ContextCL, error) {
	c := &ContextCL{}
	c.log = common.NewLogger(name)
	c.name = name
	c.KernelMutex = &sync.Mutex{}
	c.kernelSource = kernelSource

	cu.Init(0)
	c.device = cu.DeviceGet(CudaDeviceID)
	c.groupSize = c.device.Attribute(cu.MAX_THREADS_PER_BLOCK)
	c.ctx = cu.CtxCreate(cu.CTX_SCHED_AUTO, c.device)
	c.ctx.SetCurrent()

	c.module = cu.ModuleLoadData(kernelSource)
	c.fn = c.module.GetFunction("pir")

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
func (c *ContextCL) Free() error {
	c.ctx.Destroy()
	return nil
}

// Returns the working group size of this context
func (c *ContextCL) GetGroupSize() uint64 {
	return c.groupSize
}

/*********************************************
 * PRIVATE METHODS
 *********************************************/
