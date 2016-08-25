package libpdb

import (
	"github.com/ryscheng/pdb/drbg"
)

type TopicHandle struct {
	Id      uint64
	Seed1   drbg.Seed
	Seed2   drbg.Seed
	EncrKey []byte
}
