package notify

// Args notifies servers of a new snapshot
type Args struct {
	SnapshotID uint64
	Addr       string
}

// Reply acknowledges the new snapshot
type Reply struct {
	Err string
}
