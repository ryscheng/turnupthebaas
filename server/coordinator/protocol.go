package coordinator

/*************
 * PROTOCOL
 *************/

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

// CommitArgs contains a set of Writes to be committed
type CommitArgs struct {
	ID      uint64
	Bucket1 uint64
	Bucket2 uint64
	// InterestVector
}

// CommitReply acknowledges commits
type CommitReply struct {
	Err string
}
