package pir

import (
	"crypto/rand"
	"fmt"

	"github.com/privacylab/talek/common"
	"github.com/privacylab/talek/pir/pircpu"
)

// Client handles the basic functionalities of a PIR client,
// generating request vectors and combining partial responses
type Client struct {
	log  *common.Logger
	name string
}

// NewClient creates a new PIR client
func NewClient(name string) *Client {
	c := &Client{}
	c.log = common.NewLogger(name)
	c.name = name
	return c
}

// GenerateRequestVectors creates numServers requestVectors
// to retrieve data at the specified bucket
func (c *Client) GenerateRequestVectors(bucket uint64, numServers uint64, numBuckets uint64) ([][]byte, error) {
	if numServers < 2 {
		c.log.Error.Printf("GenerateRequestVectors called with too few servers=%v", numServers)
		return nil, fmt.Errorf("numServers=%v must be >1", numServers)
	}
	if bucket >= numBuckets {
		c.log.Error.Printf("GenerateRequestVectors called with invalid bucket=%v, numBuckets=%v", bucket, numBuckets)
		return nil, fmt.Errorf("bucket=%v must be <numBuckets=%v", bucket, numBuckets)
	}

	req := make([][]byte, numServers)
	numBytes := numBuckets / 8
	if (numBuckets % 8) != 0 {
		numBytes++
	}

	// Encode the secret
	req[0] = make([]byte, numBytes)
	req[0][bucket/8] |= 1 << (bucket % 8)

	var err error
	// Generate numServers-1 random request vectors
	for i := uint64(1); i < numServers; i++ {
		req[i] = make([]byte, numBytes)
		_, err = rand.Read(req[i])
		if err != nil {
			c.log.Error.Printf("GenerateRequestVectors failed: error generating random numbers %v", err)
			return nil, err
		}
		// XOR this request vector into the secret
		pircpu.XorBytes(req[0], req[0], req[i])
	}

	return req, nil
}

// CombineResponses returns the result from XORing all responses together
// Precondition: all responses are the same length
// Returns a byte array of the result
func (c *Client) CombineResponses(responses [][]byte) ([]byte, error) {
	if responses == nil || len(responses) < 1 {
		c.log.Error.Printf("CombineResponses failed: no responses input")
		return nil, fmt.Errorf("no responses input")
	}
	length := len(responses[0])
	result := make([]byte, length)
	copy(result, responses[0])

	for i := 1; i < len(responses); i++ {
		// Combine into result
		if pircpu.XorBytes(result, result, responses[i]) != length {
			c.log.Error.Printf("CombineResponses failed: malformed response %v", i)
			return nil, fmt.Errorf("malformed response %v", i)
		}
	}
	return result, nil
}
