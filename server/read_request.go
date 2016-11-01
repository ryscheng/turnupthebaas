package server

import (
	"github.com/ryscheng/pdb/common"
)

/**
 * Embodies a single read request
 * Reads require a response on the ReplyChan
 */
type ReadRequest struct {
	Args      *common.ReadArgs
	ReplyChan chan []byte
}

func (r *ReadRequest) Reply(data []byte) {
	r.ReplyChan <- data
	close(r.ReplyChan)
}

/**
 * Embodies a single batch read request
 * Reads require a response on the ReplyChan
 */
type BatchReadRequest struct {
	Args      *common.BatchReadArgs
	ReplyChan chan *common.BatchReadReply
}

func (r *BatchReadRequest) Reply(reply *common.BatchReadReply) {
	r.ReplyChan <- reply
	//close(r.ReplyChan)
}
