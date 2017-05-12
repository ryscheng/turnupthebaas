package coordinator

/*************
 * PROTOCOL
 *************/

// GetInfoReply contains general state about the server
type GetInfoReply struct {
	Err  string
	Name string
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
