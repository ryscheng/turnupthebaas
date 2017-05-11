package coordinator

/*************
 * PROTOCOL
 *************/

// GetInfoReply contains general state about the server
type GetInfoReply struct {
	Err string
}

// CommitArgs contains a set of Writes to be committed
type CommitArgs struct {
}

// CommitReply acknowledges commits
type CommitReply struct {
	Err string
}
