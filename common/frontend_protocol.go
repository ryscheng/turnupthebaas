package common

import (
	"errors"
)

// Error provides RPC errors as strings.
type Error string

/*************
 * PROTOCOL
 *************/

// PingArgs are passed in Pings.
type PingArgs struct {
	Msg string
}

// PingReply are passed in pongs.
type PingReply struct {
	Err string
	Msg string
}

// WriteArgs are passed in writes.
type WriteArgs struct {
	Bucket1        uint64
	Bucket2        uint64
	Data           []byte
	InterestVector []byte
	//Internal
	GlobalSeqNo uint64
	ReplyChan   chan *WriteReply
}

// WriteReply contain return status of writes
type WriteReply struct {
	Err         string
	GlobalSeqNo uint64
}

// PirArgs have the actual PIR for shards to perform.
type PirArgs struct {
	RequestVector []byte
	PadSeed       []byte
}

// ReadArgs have the ReadArgs for each trust domain in unencrypted form.
type ReadArgs struct {
	TD []PirArgs
}

// EncodedReadArgs are a trust-domain-encrypted form of ReadArgs
type EncodedReadArgs struct {
	ClientKey [32]byte
	Nonce     [24]byte
	PirArgs   [][]byte //An encrypted PirArgs for each trust domain
}

// ReadReply contain the response to a read.
type ReadReply struct {
	Err         string
	Data        []byte
	GlobalSeqNo Range
}

// Combine xors two partial read replies together
func (r *ReadReply) Combine(other []byte) error {
	if len(r.Data) != len(other) {
		return errors.New("cannot combine responses of different length")
	}
	for i := 0; i < len(r.Data); i++ {
		r.Data[i] = r.Data[i] ^ other[i]
	}
	return nil
}

// ReadRequest is the actual request sent to the frontend from libtalek.
// response occurs on the provided replychan
type ReadRequest struct {
	Args      *EncodedReadArgs
	ReplyChan chan *ReadReply
}

// Reply returns the response to the client.
func (r *ReadRequest) Reply(reply *ReadReply) {
	r.ReplyChan <- reply
	close(r.ReplyChan)
}

// GetUpdatesArgs is the empty pointer fullfilling the RPC calling convention.
type GetUpdatesArgs struct {
}

// GetUpdatesReply has the interestvector response for a getupdates call
type GetUpdatesReply struct {
	Err            string
	InterestVector []byte
}
