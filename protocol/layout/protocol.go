package layout

/**********************
 * PROTOCOL
 **********************/

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
