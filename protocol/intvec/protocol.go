package intvec

/**********************
 * PROTOCOL
 **********************/

// Args requests the global interest vector
type Args struct {
	SnapshotID uint64
}

// Reply returns the global interest vector
type Reply struct {
	Err        string
	SnapshotID uint64
	IntVec     []uint64 // Serialization of bloom filter
}
