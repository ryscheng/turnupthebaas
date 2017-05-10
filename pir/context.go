package pir

// Context abstracts out the common interface for ContextCL and ContextCUDA
// A Context represents a specific computing context, a kernel, and set of devices.
type Context interface {
	GetGroupSize() int
	Free() error
}
