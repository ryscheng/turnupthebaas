package common

import (
	"errors"
)

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
	ReplyChan   chan *WriteReply
}

type WriteReply struct {
	Err         string
	GlobalSeqNo uint64
}

type PirArgs struct {
	RequestVector []byte
	PadSeed       []byte
}

type ReadArgs struct {
	TD []PirArgs
}

type EncodedReadArgs struct {
	ClientKey [32]byte
	Nonce     [24]byte
	PirArgs   [][]byte //An encrypted PirArgs for each trust domain
}

type ReadReply struct {
	Err         string
	Data        []byte
	GlobalSeqNo Range
}

func (r *ReadReply) Combine(other []byte) error {
	if len(r.Data) != len(other) {
		return errors.New("Cannot combine responses of different length.")
	}
	for i := 0; i < len(r.Data); i++ {
		r.Data[i] = r.Data[i] ^ other[i]
	}
	return nil
}

/**
 * Embodies a single read request
 * Reads require a response on the ReplyChan
 */
type ReadRequest struct {
	Args      *ReadArgs
	ReplyChan chan *ReadReply
}

func (r *ReadRequest) Reply(reply *ReadReply) {
	r.ReplyChan <- reply
	close(r.ReplyChan)
}

type GetUpdatesArgs struct {
}

type GetUpdatesReply struct {
	Err            string
	InterestVector []byte
}
