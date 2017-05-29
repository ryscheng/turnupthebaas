package coordinator

/*************
 * PROTOCOL
 *************/

// IDSize is the size of unique IDs in bytes
// @todo for some reason uint64's require 10 bytes?
const IDSize = 10

// GetInfoReply contains general state about the server
type GetInfoReply struct {
	Err        string
	Name       string
	SnapshotID uint64
}

// GetLayoutArgs requests the layout for a shard
type GetLayoutArgs struct {
	SnapshotID uint64
	ShardID    uint64
	NumShards  uint64
}

// GetLayoutReply returns the layout for a shard
type GetLayoutReply struct {
	Err        string
	SnapshotID uint64
	Layout     []uint64
}

// GetIntVecArgs requests the global interest vector
type GetIntVecArgs struct {
	SnapshotID uint64
}

// GetIntVecReply returns the global interest vector
type GetIntVecReply struct {
	Err        string
	SnapshotID uint64
	IntVec     []uint64 // Serialization of bloom filter
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
