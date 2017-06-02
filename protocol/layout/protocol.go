package layout

/**********************
 * PROTOCOL
 **********************/

// Args requests the layout for a shard
type Args struct {
	SnapshotID uint64
	Index      uint64
	NumSplit   uint64
}

// Reply returns the layout for a shard
type Reply struct {
	Err        string
	SnapshotID uint64
	Layout     []uint64
}
