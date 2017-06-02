package pir

import (
	"encoding/binary"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("test")
	if c == nil {
		t.Errorf("Client should not be nil")
	}
}

func TestGenerateRequestVectors(t *testing.T) {
	c := NewClient("test")
	reqVec, err := c.GenerateRequestVectors(1, 3, 64)
	if err != nil {
		t.Errorf("GenerateRequestVectors failed: %v", err)
	}
	if len(reqVec) != 3 {
		t.Errorf("GenerateRequestVectors produced too few request vectors %v, expected 3", len(reqVec))
	}
	resultBytes, err := c.CombineResponses(reqVec)
	if err != nil {
		t.Errorf("CombineResponses failed: %v", err)
	}
	result, _ := binary.Uvarint(resultBytes)
	if result != 2 {
		t.Errorf("Secret request vector should translate to 2, not %v", result)
	}

}

func TestGenerateRequestVectorsInvalidNumServers(t *testing.T) {
}

func TestGenerateRequestVectorsInvalidBucket(t *testing.T) {
}

func TestCombineResponses(t *testing.T) {
}

func TestCombineResponsesNone(t *testing.T) {
}

func TestCombineResponsesInvalid(t *testing.T) {
}
