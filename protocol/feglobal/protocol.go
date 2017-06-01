package feglobal

import (
	"github.com/privacylab/talek/protocol/fedomain"
)

/*************
 * PROTOCOL
 *************/

// GetInfoReply contains general state about the server
type GetInfoReply struct {
	Err        string
	Name       string
	SnapshotID uint64
}

// WriteArgs are passed in writes.
type WriteArgs struct {
	ID             uint64
	Bucket1        uint64
	Bucket2        uint64
	Data           []byte
	InterestVector []uint64
}

// WriteReply contain return status of writes
type WriteReply struct {
	Err string
}

// ReadArgs are a trust-domain-encrypted form of ReadArgs
type ReadArgs struct {
	TD []fedomain.EncPIRArgs
}

// ReadReply contain the response to a read.
type ReadReply struct {
	Err  string
	Data []byte
}
