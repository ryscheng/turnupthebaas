package common

type BatchReadArgs struct {
	Args       []ReadArgs // Set of Read requests
	SeqNoRange Range
}

type BatchReadReply struct {
	Err     string
	Replies []ReadReply
}

/*************
 * OTHER TYPES
 *************/

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
