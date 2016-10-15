package common

type Error string

/*************
 * PROTOCOL
 *************/

type PingArgs struct {
	Msg string
}

type PingReply struct {
	Err string
	Msg string
}

type WriteArgs struct {
	Bucket1        uint32
	Bucket2        uint32
	Data           []byte
	InterestVector []byte
	//Internal
	GlobalSeqNo uint64
}

type WriteReply struct {
	Err string
}

type PIRArgs struct {
	RequestVector []byte
	PadSeed       []byte
}

type ReadArgs struct {
	ForTd []PIRArgs // Set of args for each trust domain
}

type ReadReply struct {
	Err  string
	Data []byte
}

type GetUpdatesArgs struct {
}

type GetUpdatesReply struct {
	Err            string
	InterestVector []byte
}
