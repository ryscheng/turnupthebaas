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

// PIRArgs for a PIR read to a specific range of shards
type PIRArgs struct {
	SnapshotID    uint64 // Snapshot to target
	ShardStart    uint64 // inclusive
	ShardEnd      uint64 // exclusive
	RequestVector []byte // PIR request vector
}

// PIRReply for a PIR read to a specific range of shards
type PIRReply struct {
	Err        string
	SnapshotID uint64 // SnapshotID of replica
	Data       []byte // result
}
