package cuckoo

import (
	"fmt"
	"strings"
	"testing"
)

// Test identical state after operations
// Test contains after insert/remove sequence
// Test insert same value twice

type Value string

func (v Value) Compare(other Comparable) int {
	otherStr := string(other.(Value))
	return strings.Compare(string(v), otherStr)
}

func TestContains(t *testing.T) {
	table := NewTable("t", 10, 10, 0)

	fmt.Printf("TestContains: Check empty ...\n")
	if table.Contains(0, 1, Value("")) {
		t.Fatalf("Empty table returned true for Contains()")
	}
}
