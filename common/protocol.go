package common

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
	for i := range r.aborted {
		if r.aborted[i] != b.aborted[i] {
			return false
		}
	}
	return true
}

type PingArgs struct {
	Msg string
}

type PingReply struct {
	Err Error
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
	Err Error
}

type ReadArgs struct {
	RequestVector []byte
}

type ReadReply struct {
	Err  Error
	Data []byte
}

type GetUpdatesArgs struct {
}

type GetUpdatesReply struct {
	Err            Error
	InterestVector []byte
}
