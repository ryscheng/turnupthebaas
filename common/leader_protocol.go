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
	Bucket1        uint64
	Bucket2        uint64
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

func (r *ReadReply) Combine(other []byte) {
	var length int
	if len(r.Data) < len(other) {
		length = len(r.Data)
	} else {
		length = len(other)
	}
	for i := 0; i < length; i++ {
		r.Data[i] = r.Data[i] ^ other[i]
	}
}

type GetUpdatesArgs struct {
}

type GetUpdatesReply struct {
	Err            string
	InterestVector []byte
}
