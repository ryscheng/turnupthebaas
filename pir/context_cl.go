//+build !travis

package pir

import (
	"fmt"
	"github.com/go-gl/cl/v1.2/cl"
)

// ContextCL represents a single OpenCL context
// Currently, we only support 1 live ShardCL per ContextCL
// Before creating a new ShardCL, the old one must be Free()
type ContextCL struct {
	platformID cl.PlatformID
	deviceID   cl.DeviceId
}

// KernelSource is the source code of the program we're going to run.
var KernelSource = `
__kernel void pir(
   __global float* input,
   __global float* output,
   const unsigned int count)
{
   int i = get_global_id(0);
   if(i < count)
     output[i] = input[i] * input[i];
}` + "\x00"

/*********************************************
 * PRIVATE METHODS
 *********************************************/
func (c *ContextCL) initCL() error {
	// Get Platform
	ids := make([]cl.PlatformID, 100)
	count := uint32(0)
	err := cl.GetPlatformIDs(uint32(len(ids)), &ids[0], &count)
	if err != cl.SUCCESS || count < 1 {
		return fmt.Errorf("NewContextCl: failed to retrieve OpenCL platform ID\n")
	}
	c.platformID = ids[0]

	// Get Device
	devices := make([]cl.DeviceId, 100)
	err = cl.GetDeviceIDs(c.platformID, cl.DEVICE_TYPE_GPU, 1, &devices[0], &count)
	if err != cl.SUCCESS || count < 1 {
		return fmt.Errorf("NewContextCl: failed to create OpenCL device group\n")
	}
	c.deviceID = devices[0]
	return nil

}
