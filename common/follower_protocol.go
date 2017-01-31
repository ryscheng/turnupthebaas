package common

type BatchReadRequest struct {
	Args       []PirArgs // Set of Read requests
	SeqNoRange Range
	RandSeed   int64
	ReplyChan chan *BatchReadReply
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

func (r *Range) Contains(val uint64) bool {
	if val < r.Start {
		return false
	}
	if val >= r.End {
		return false
	}
	for _, elt := range r.Aborted {
		if val == elt {
			return false
		}
	}
	return true
}
