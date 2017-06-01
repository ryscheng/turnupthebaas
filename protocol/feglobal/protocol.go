package feglobal

/*************
 * PROTOCOL
 *************/

// GetInfoReply contains general state about the server
type GetInfoReply struct {
	Err        string
	Name       string
	SnapshotID uint64
}

// WriteArgs are passed in writes.
type WriteArgs struct {
	ID             uint64
	Bucket1        uint64
	Bucket2        uint64
	Data           []byte
	InterestVector []uint64
}

// WriteReply contain return status of writes
type WriteReply struct {
	Err string
}

// ReadArgs are a trust-domain-encrypted form of ReadArgs
type ReadArgs struct {
	TD []EncPIRArgs
}

// ReadReply contain the response to a read.
type ReadReply struct {
	Err        string
	SnapshotID uint64 // SnapshotID of replica
	Data       []byte
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
