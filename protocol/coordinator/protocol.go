package coordinator

/**********************
 * PROTOCOL
 **********************/

// IDSize is the size of unique IDs in bytes
// @todo for some reason uint64's require 10 bytes?
const IDSize = 10

// GetInfoReply contains general state about the server
type GetInfoReply struct {
	Err        string
	Name       string
	SnapshotID uint64
}

// CommitArgs contains a set of Writes to be committed
type CommitArgs struct {
	ID        uint64
	Bucket1   uint64
	Bucket2   uint64
	IntVecLoc []uint64 // Represents the hash locations
}

// CommitReply acknowledges commits
type CommitReply struct {
	Err string
}
