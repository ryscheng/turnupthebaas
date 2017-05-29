package replica

/*************
 * PROTOCOL
 *************/

// GetInfoReply contains general state about the server
type GetInfoReply struct {
	Err        string
	Name       string
	SnapshotID uint64
}

// ReadArgs for a batch PIR read to a specific range of shards
type ReadArgs struct {
	SnapshotID     uint64 // Snapshot to target
	ShardStart     uint64 // inclusive
	ShardEnd       uint64 // exclusive
	RequestVectors []byte // Batch of PIR requests
}

// ReadReply for a batch PIR read to a specific range of shards
type ReadReply struct {
	Err        string
	SnapshotID uint64 // SnapshotID of replica
	Data       []byte // result
}
