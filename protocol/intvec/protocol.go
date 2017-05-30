package intvec

/**********************
 * PROTOCOL
 **********************/

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
