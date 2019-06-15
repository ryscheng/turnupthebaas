package common

// ReplicaWriteArgs forwards a client write from frontend to replicas.
type ReplicaWriteArgs struct {
	WriteArgs
	EpochFlag    bool
	InterestFlag bool
}

// ReplicaWriteReply contain return status of writes
type ReplicaWriteReply struct {
	Err         string
	GlobalSeqNo uint64
	InterestVec []byte
	Signature   []byte
}

// BatchReadRequest are a batch of requests sent to PIR servers from frontend.
type BatchReadRequest struct {
	Args       []EncodedReadArgs // Set of Read requests
	SeqNoRange Range
	ReplyChan  chan *BatchReadReply `json:"-"`
}

// BatchReadReply is a response to a BatchReadRequest.
type BatchReadReply struct {
	Err     string
	Replies []ReadReply
}

/*************
 * OTHER TYPES
 *************/

// Range is a range of sequence numbers
type Range struct {
	Start   uint64 //inclusive
	End     uint64 //exclusive
	Aborted []uint64
}

// Equals compares two ranges.
func (r *Range) Equals(b Range) bool {
	if r.Start != b.Start || r.End != b.End || len(r.Aborted) != len(b.Aborted) {
		return false
	}
	for i := range r.Aborted {
		if r.Aborted[i] != b.Aborted[i] {
			return false
		}
	}
	return true
}

// Contains checks subset inclusion of a range
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
