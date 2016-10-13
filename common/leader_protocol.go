package common

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

/*************
 * OTHER TYPES
 *************/

type Error string

type Range struct {
	start   uint64 //inclusive
	end     uint64 //exclusive
	aborted []uint64
}

func (r *Range) Equals(b Range) bool {
	if r.start != b.start || r.end != b.end || len(r.aborted) != len(b.aborted) {
		return false
	}
	for i, _ := range r.aborted {
		if r.aborted[i] != b.aborted[i] {
			return false
		}
	}
	return true
}
