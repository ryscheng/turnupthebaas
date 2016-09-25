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
	var b1, b2, evictb1, evictb2 int
	var evictVal Comparable
	var val Value

	// Insert random values until we've reached a limit
	fmt.Printf("TestFullTable: Insert until reach capacity...\n")
	for ok {
		b1 = randBucket(numBuckets)
		b2 = randBucket(numBuckets)
		val = Value(strconv.Itoa(rand.Int()))
		entries = append(entries, Entry{b1, b2, val})
		evictb1, evictb2, evictVal, ok = table.Insert(b1, b2, val)

		if ok {
			count += 1
			if table.Contains(b1, b2, val) == false {
				t.Fatalf("Insert() succeeded, but Contains failed\n")
			}
			if count != table.GetNumElements() {
				t.Fatalf("Number of successful Inserts(), %v, does not match GetNumElements(), %v \n", count, table.GetNumElements())
			}
		}
	}

	// Middle count check
	fmt.Printf("TestFullTable: Mid-count check...\n")
	if count != table.GetNumElements() {
		t.Fatalf("Number of successful Inserts(), %v, does not match GetNumElements(), %v \n", count, table.GetNumElements())
	}

	// Remove elements one by one
	fmt.Printf("TestFullTable: Remove each element...\n")
	for _, entry := range entries {
		if entry.Bucket1 != evictb1 || entry.Bucket2 != evictb2 || entry.Data.Compare(evictVal) != 0 {
			ok = table.Remove(entry.Bucket1, entry.Bucket2, entry.Data)
			if ok == false {
				t.Fatalf("Cannot Remove() an element believed to be in the table")
			} else {
				count -= 1
				if count != table.GetNumElements() {
					t.Fatalf("GetNumElements()=%v returned a value that didn't match what was expected=%v \n", table.GetNumElements(), count)
				}
			}
		}
	}

	// Final count check
	fmt.Printf("TestFullTable: Final count check...\n")
	if table.GetNumElements() != 0 {
		t.Fatalf("GetNumElements() returns %v when table should be empty \n", table.GetNumElements())
	}

	fmt.Printf("... done\n")
}

func TestDuplicateValues(t *testing.T) {
	fmt.Printf("TestDuplicateValues: ...\n")
	table := NewTable("t", 10, 2, 0)

	eb1, eb2, ev, ok := table.Insert(0, 1, Value("v"))
	if eb1 != -1 || eb2 != -1 || ev != nil || ok == false {
		t.Fatalf("Error inserting value \n")
	}

	eb1, eb2, ev, ok = table.Insert(0, 1, Value("v"))
	if eb1 != -1 || eb2 != -1 || ev != nil || ok == false {
		t.Fatalf("Error inserting value again \n")
	}

	eb1, eb2, ev, ok = table.Insert(1, 2, Value("v"))
	if eb1 != -1 || eb2 != -1 || ev != nil || ok == false {
		t.Fatalf("Error inserting value in shifted buckets\n")
	}

	if table.Remove(0, 1, Value("v")) == false {
		t.Fatalf("Error removing value 1st time\n")
	}

	if table.Remove(0, 1, Value("v")) == false {
		t.Fatalf("Error removing value 2nd time\n")
	}

	if table.Remove(1, 2, Value("v")) == false {
		t.Fatalf("Error removing value 3rd time\n")
	}

	fmt.Printf("... done\n")
}

func TestLoadFactor(t *testing.T) {
	fmt.Printf("TestLoadFactor: ...\n")
	numBuckets := 1000
	var table *Table
	var val Value

	for depth := 1; depth < 32; depth *= 2 {
		count := -1
		ok := true
		table = NewTable("t", numBuckets, depth, int64(depth))
		for ok {
			count += 1
			val = Value(strconv.Itoa(rand.Int()))
			_, _, _, ok = table.Insert(randBucket(numBuckets), randBucket(numBuckets), val)
		}

		if table.GetNumElements() != count {
			t.Fatalf("Mismatch count=%v with GetNumElements()=%v \n", count, table.GetNumElements())
		}
		fmt.Printf("count=%v, capacity=%v, loadfactor=%v \n", count, table.GetCapacity(), (float64(count) / float64(table.GetCapacity())))
	}

	fmt.Printf("... done\n")
}
