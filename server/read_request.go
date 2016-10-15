package server

import (
	"github.com/ryscheng/pdb/common"
)

/**
 * Embodies a single read request
 * Reads require a response on the readReplyChan
 */
type ReadRequest struct {
	Args      *common.ReadArgs
	ReplyChan chan []byte
}

func (r *ReadRequest) ReplyRead(data []byte) {
	r.ReplyChan <- data
	close(r.ReplyChan)
}
