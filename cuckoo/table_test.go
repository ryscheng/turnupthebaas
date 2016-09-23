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

type Entry struct {
	Bucket1 int
	Bucket2 int
	Data    Value
}

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

func TestGetCapacity(t *testing.T) {
	fmt.Printf("TestGetCapacity: ...\n")

	table := NewTable("t", 10, 2, 0)
	if table.GetCapacity() != 20 {
		t.Fatalf("table returned wrong value for GetCapacity=%v. Expecting 20\n")
	}

	table = NewTable("t", 1, 1, 0)
	if table.GetCapacity() != 1 {
		t.Fatalf("table returned wrong value for GetCapacity=%v. Expecting 1\n")
	}

	table = NewTable("t", 0, 0, 0)
	if table.GetCapacity() != 0 {
		t.Fatalf("table returned wrong value for GetCapacity=%v. Expecting 0\n")
	}

	fmt.Printf("... done \n")
}

func TestBasic(t *testing.T) {
	table := NewTable("t", 10, 2, 0)

	fmt.Printf("TestBasic: check empty...\n")
	if table.GetNumElements() != 0 {
		t.Fatalf("empty table returned %v for GetNumElements()\n", table.GetNumElements())
	}

	fmt.Printf("TestBasic: Check contains non-existent value...\n")
	if table.Contains(0, 1, Value("")) == true {
		t.Fatalf("empty table returned true for Contains()\n")
	}

	fmt.Printf("TestBasic: remove non-existent value ...\n")
	if table.Remove(0, 1, Value("value1")) == true {
		t.Fatalf("empty table returned true for Remove()\n")
	}

	fmt.Printf("TestBasic: check empty...\n")
	if table.GetNumElements() != 0 {
		t.Fatalf("empty table returned %v for GetNumElements()\n", table.GetNumElements())
	}

	fmt.Printf("TestBasic: Insert value ...\n")
	eb1, eb2, v, ok := table.Insert(0, 1, Value("value1"))
	if eb1 != -1 || eb2 != -1 || v != nil || ok != true {
		t.Fatalf("error inserting into table (0, 1, value1)\n")
	}

	fmt.Printf("TestBasic: Check inserted value...\n")
	if table.Contains(0, 1, Value("value1")) == false {
		t.Fatalf("cannot find recently inserted value\n")
	}

	fmt.Printf("TestBasic: Check non-existent value...\n")
	if table.Contains(0, 1, Value("value2")) == true {
		t.Fatalf("contains a non-existent value\n")
	}

	fmt.Printf("TestBasic: check 1 element...\n")
	if table.GetNumElements() != 1 {
		t.Fatalf("empty table returned %v for GetNumElements()\n", table.GetNumElements())
	}

	fmt.Printf("TestBasic: remove existing value ...\n")
	if table.Remove(0, 1, Value("value1")) == false {
		t.Fatalf("error removing existing value (0, 1, value1)\n")
	}

	fmt.Printf("TestBasic: check 0 element...\n")
	if table.GetNumElements() != 0 {
		t.Fatalf("empty table returned %v for GetNumElements()\n", table.GetNumElements())
	}

	fmt.Printf("TestBasic: remove recently removed value ...\n")
	if table.Remove(0, 1, Value("value1")) == true {
		t.Fatalf("empty table returned true for Remove()\n")
	}

	fmt.Printf("TestBasic: check 0 element...\n")
	if table.GetNumElements() != 0 {
		t.Fatalf("empty table returned %v for GetNumElements()\n", table.GetNumElements())
	}

	fmt.Printf("... done \n")
}

func TestOutOfBounds(t *testing.T) {
	table := NewTable("t", 10, 2, 0)

	fmt.Printf("TestOutOfBounds: Insert() out of bounds...\n")
	eb1, eb2, v, ok := table.Insert(100, 100, Value("value1"))
	if ok == true {
		t.Fatalf("Insert returned true with out of bound buckets\n")
	}
	if eb1 != -1 || eb2 != -1 || v != nil {
		t.Fatalf("Insert returned wrong values\n")
	}

	fmt.Printf("TestOutOfBounds: Contains() out of bounds...\n")
	if table.Contains(100, 100, Value("value1")) == true {
		t.Fatalf("Contains returned true with out of bound buckets\n")
	}

	fmt.Printf("TestOutOfBounds: Remove() out of bounds...\n")
	if table.Remove(100, 100, Value("value1")) == true {
		t.Fatalf("Remove() returned true with out of bound buckets\n")
	}

	fmt.Printf("... done \n")
}

func TestFullTable(t *testing.T) {
	numBuckets := 100
	depth := 4

	capacity := numBuckets * depth
	entries := make([]Entry, 0, capacity)
	table := NewTable("t", numBuckets, depth, 0)
	ok := true
	count := 0
	var b1, b2 int
	var compVal Comparable
	var val Value

	for ok {
		b1 = randBucket(numBuckets)
		b2 = randBucket(numBuckets)
		val = Value(strconv.Itoa(rand.Int()))
		entries = append(entries, Entry{b1, b2, val})
		b1, b2, compVal, ok = table.Insert(b1, b2, val)
		count += 1
	}

	fmt.Println(compVal)
	fmt.Println(count)
	fmt.Printf("TestInsert: Check empty ...\n")
	fmt.Printf("... done\n")
}

func TestLoadFactor(t *testing.T) {
}
