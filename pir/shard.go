package pir

import (
	"github.com/willf/bitset"
)

// Shard abstracts out the common interface for ShardCPU, ShardCUDA, and ShardOpenCL
// Each are backed by a different PIR implementation
type Shard interface {
	Free() error
	GetName() string
	GetBucketSize() int
	GetNumBuckets() int
	GetData() []byte
	Read(reqs []*bitset.BitSet) ([]byte, error)
}
