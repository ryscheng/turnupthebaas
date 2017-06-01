package fedomain

/*************
 * PROTOCOL
 *************/

// GetInfoReply contains general state about the server
type GetInfoReply struct {
	Err        string
	Name       string
	SnapshotID uint64
}
