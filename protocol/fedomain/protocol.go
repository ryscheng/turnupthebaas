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

// EncPIRArgs contains the encrypted PIR parameters for a single trust domain
type EncPIRArgs struct {
	SnapshotID uint64 // Snapshot to target
	ShardStart uint64 // inclusive
	ShardEnd   uint64 // exclusive
	ClientKey  [32]byte
	Nonce      [24]byte
	PirArgs    []byte // Encrypted PIR arg batch
}

// EncPIRReply contains the encrypted PIR reply from a single trust domain
type EncPIRReply struct {
	Err        string
	SnapshotID uint64 // SnapshotID of replica
	Data       []byte // result
}
