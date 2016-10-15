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
	Start   uint64 //inclusive
	End     uint64 //exclusive
	Aborted []uint64
}

func (r *Range) Equals(b Range) bool {
	if r.Start != b.Start || r.End != b.End || len(r.Aborted) != len(b.Aborted) {
		return false
	}
	for i, _ := range r.Aborted {
		if r.Aborted[i] != b.Aborted[i] {
			return false
		}
	}
	return true
}
