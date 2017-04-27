package pir

import (
	"github.com/willf/bitset"
)

type Shard interface {
	Free() error
	GetName() string
	GetBucketSize() int
	GetNumBuckets() int
	GetData() []byte
	Read(reqs []bitset.BitSet) ([]byte, error)
}
