package cuckoo

import (
	"fmt"
	"math/rand"
	"strconv"
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

func randBucket(numBuckets int) int {
	result := rand.Int() % numBuckets
	if result < 0 {
		result = result * -1
	}
	return result
}

func TestContains(t *testing.T) {
	table := NewTable("t", 10, 2, 0)

	fmt.Printf("TestContains: Check empty ...\n")
	if table.Contains(0, 1, Value("")) == true {
		t.Fatalf("empty table returned true for Contains()")
	}

	fmt.Printf("TestContains: Insert value ...\n")
	eb1, eb2, v, ok := table.Insert(0, 1, Value("value1"))
	if eb1 != -1 || eb2 != -1 || v != nil || ok != true {
		t.Fatalf("error inserting into table (0, 1, value1)")
	}

	fmt.Printf("TestContains: Check inserted value...\n")
	if table.Contains(0, 1, Value("value1")) == false {
		t.Fatalf("cannot find recently inserted value")
	}

	fmt.Printf("TestContains: Check non-existent value...\n")
	if table.Contains(0, 1, Value("value2")) == true {
		t.Fatalf("contains a non-existent value")
	}

	fmt.Printf("TestContains: Check out of bounds...\n")
	if table.Contains(100, 100, Value("value1")) == true {
		t.Fatalf("contains returned true with out of bound buckets")
	}

	fmt.Printf("... done\n")
}

func TestInsertRemove(t *testing.T) {
	table := NewTable("t", 10, 2, 0)

	fmt.Printf("TestInsertRemove: \n")
}

func TestFullTable(t *testing.T) {
	numBuckets := 100
	depth := 4
	table := NewTable("t", numBuckets, depth, 0)
	var b1, b2 int
	var val Value

	for i := 0; i < 385; i++ {
		b1 = randBucket(numBuckets)
		b2 = randBucket(numBuckets)
		val = Value(strconv.Itoa(rand.Int()))
		table.Insert(b1, b2, val)
	}

	fmt.Printf("TestInsert: Check empty ...\n")
	fmt.Printf("... done\n")
}

func TestLoadFactor(t *testing.T) {
}
