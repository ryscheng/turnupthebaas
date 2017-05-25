package coordinator

// NotifyArgs notifies servers of a new snapshot
type NotifyArgs struct {
	SnapshotID uint64
}

// NotifyReply acknowledges the new snapshot
type NotifyReply struct {
	Err string
}
