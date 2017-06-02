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
	c := NewClient("test")
	_, err := c.GenerateRequestVectors(1, 1, 64)
	if err == nil {
		t.Errorf("GenerateRequestVectors should fail with 1 server")
	}
}

func TestGenerateRequestVectorsInvalidBucket(t *testing.T) {
	c := NewClient("test")
	_, err := c.GenerateRequestVectors(65, 3, 64)
	if err == nil {
		t.Errorf("GenerateRequestVectors should fail with out of bounds bucket")
	}
}

func TestCombineResponses(t *testing.T) {
	c := NewClient("test")
	result, err := c.CombineResponses([][]byte{
		[]byte{1, 2, 3, 4, 5},
		[]byte{1, 2, 3, 4, 5},
	})
	if err != nil {
		t.Errorf("CombineResponses shouldn't have failed")
	}
	for _, b := range result {
		if b != 0 {
			t.Errorf("CombineResponses should return 0")
		}
	}
}

func TestCombineResponsesNone(t *testing.T) {
	c := NewClient("test")
	_, err := c.CombineResponses(make([][]byte, 0))
	if err == nil {
		t.Errorf("CombineResponses should have failed with no responses to combine")
	}
}

func TestCombineResponsesInvalid(t *testing.T) {
	c := NewClient("test")
	_, err := c.CombineResponses([][]byte{
		[]byte{1, 2, 3},
		[]byte{1},
	})
	if err == nil {
		t.Errorf("CombineResponses should have failed with mismatched responses")
	}
	_, err = c.CombineResponses([][]byte{
		[]byte{1},
		[]byte{1, 2, 3},
	})
	if err != nil {
		t.Errorf("CombineResponses is okay with bigger later responses")
	}
}
