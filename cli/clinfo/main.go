package main

import (
	"log"
	"unsafe"

	"github.com/go-gl/cl/v1.2/cl"
)

const (
	// DataSize is the size of the buffer for the string we're going to get back from the CL device.
	DataSize = 1024
)

var PlatformProperties = map[string]cl.PlatformInfo{
	"PLATFORM_NAME":       cl.PLATFORM_NAME,
	"PLATFORM_VENDOR":     cl.PLATFORM_VENDOR,
	"PLATFORM_VERSION":    cl.PLATFORM_VERSION,
	"PLATFORM_PROFILE":    cl.PLATFORM_PROFILE,
	"PLATFORM_EXTENSIONS": cl.PLATFORM_EXTENSIONS,
}
var DeviceProperties = map[string]cl.DeviceInfo{
	"DEVICE_NAME":                        cl.DEVICE_NAME,
	"DEVICE_VENDOR":                      cl.DEVICE_VENDOR,
	"DEVICE_VERSION":                     cl.DEVICE_VERSION,
	"DEVICE_PROFILE":                     cl.DEVICE_PROFILE,
	"DEVICE_MAX_COMPUTE_UNITS":           cl.DEVICE_MAX_COMPUTE_UNITS,
	"DEVICE_GLOBAL_MEM_SIZE":             cl.DEVICE_GLOBAL_MEM_SIZE,
	"DEVICE_LOCAL_MEM_SIZE":              cl.DEVICE_LOCAL_MEM_SIZE,
	"DEVICE_PREFERRED_VECTOR_WIDTH_CHAR": cl.DEVICE_PREFERRED_VECTOR_WIDTH_CHAR,
	// "DEVICE_EXTENSIONS": cl.DEVICE_EXTENSIONS,
}

// StatInfo prints a bunch of useful information about the currently available CL devices
func StatInfo() {
	// Get platforms
	ids := make([]cl.PlatformID, 100)
	numPlatforms := uint32(0)
	cl.GetPlatformIDs(uint32(len(ids)), &ids[0], &numPlatforms)
	for i := 0; i < int(numPlatforms); i++ {
		data := make([]byte, DataSize)
		size := uint64(0)
		// Platform Info
		for propName, propVal := range PlatformProperties {
			cl.GetPlatformInfo(ids[i], propVal, DataSize, unsafe.Pointer(&data[0]), &size)
			str := string(data[0:size])
			log.Printf("%v:\t\t %v\n", propName, str)
		}
		// Device Info
		devices := make([]cl.DeviceId, 100)
		numDevices := uint32(0)
		cl.GetDeviceIDs(ids[i], cl.DEVICE_TYPE_ALL, uint32(len(devices)), &devices[0], &numDevices)
		log.Printf("---\n")
		log.Println("Devices: ")
		for y := 0; y < int(numDevices); y++ {
			log.Printf("DeviceIdAddr:%v\n", &devices[y])
			for propName, propVal := range DeviceProperties {
				cl.GetDeviceInfo(devices[y], propVal, DataSize, unsafe.Pointer(&data[0]), &size)
				str := string(data[0:size])
				log.Printf("\t %v: %v", propName, str)
			}
			log.Printf("---\n")
		}
	}
}

func main() {
	StatInfo()
}
