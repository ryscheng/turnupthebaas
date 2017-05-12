package cuckoo

import (
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
)

// Test identical state after operations
// Test contains after insert/remove sequence
// Test insert same value twice

func GetBytes(val string) []byte {
	buf := make([]byte, 64)
	copy(buf, bytes.NewBufferString(val).Bytes())
	return buf
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

	table := NewTable("t", 10, 2, 64, nil, 0)
	if table.GetCapacity() != 20 {
		t.Fatalf("table returned wrong value for GetCapacity=%v. Expecting 20\n", table.GetCapacity())
	}

	table = NewTable("t", 1, 1, 64, nil, 0)
	if table.GetCapacity() != 1 {
		t.Fatalf("table returned wrong value for GetCapacity=%v. Expecting 1\n", table.GetCapacity())
	}

	table = NewTable("t", 0, 0, 64, nil, 0)
	if table.GetCapacity() != 0 {
		t.Fatalf("table returned wrong value for GetCapacity=%v. Expecting 0\n", table.GetCapacity())
	}

	fmt.Printf("... done \n")
}

func TestBasic(t *testing.T) {
	table := NewTable("t", 10, 2, 64, nil, 0)

	fmt.Printf("TestBasic: check empty...\n")
	if table.GetNumElements() != 0 {
		t.Fatalf("empty table returned %v for GetNumElements()\n", table.GetNumElements())
	}

	fmt.Printf("TestBasic: Check contains non-existent value...\n")
	if table.Contains(&Item{0, GetBytes(""), 0, 1}) {
		t.Fatalf("empty table returned true for Contains()\n")
	}

	fmt.Printf("TestBasic: remove non-existent value ...\n")
	if table.Remove(&Item{1, GetBytes("value1"), 0, 1}) {
		t.Fatalf("empty table returned true for Remove()\n")
	}

	fmt.Printf("TestBasic: check empty...\n")
	if table.GetNumElements() != 0 {
		t.Fatalf("empty table returned %v for GetNumElements()\n", table.GetNumElements())
	}

	fmt.Printf("TestBasic: Insert value ...\n")
	ok, itm := table.Insert(&Item{1, GetBytes("value1"), 0, 1})
	if itm != nil || !ok {
		t.Fatalf("error inserting into table (0, 1, value1)\n")
	}

	fmt.Printf("TestBasic: Check inserted value...\n")
	if !table.Contains(&Item{1, GetBytes("value1"), 0, 1}) {
		t.Fatalf("cannot find recently inserted value\n")
	}

	fmt.Printf("TestBasic: Check inserted value w/o full reference...\n")
	if !table.Contains(&Item{1, GetBytes(""), 0, 1}) {
		t.Fatalf("cannot find recently inserted value\n")
	}

	fmt.Printf("TestBasic: Check non-existent value...\n")
	if table.Contains(&Item{2, GetBytes("value2"), 0, 1}) {
		t.Fatalf("contains a non-existent value\n")
	}

	fmt.Printf("TestBasic: check 1 element...\n")
	if table.GetNumElements() != 1 {
		t.Fatalf("empty table returned %v for GetNumElements()\n", table.GetNumElements())
	}

	fmt.Printf("TestBasic: remove existing value ...\n")
	if !table.Remove(&Item{1, GetBytes("value1"), 0, 1}) {
		t.Fatalf("error removing existing value (0, 1, value1)\n")
	}

	fmt.Printf("TestBasic: check 0 element...\n")
	if table.GetNumElements() != 0 {
		t.Fatalf("empty table returned %v for GetNumElements()\n", table.GetNumElements())
	}

	fmt.Printf("TestBasic: remove recently removed value ...\n")
	if table.Remove(&Item{1, GetBytes("value1"), 0, 1}) {
		t.Fatalf("empty table returned true for Remove()\n")
	}

	fmt.Printf("TestBasic: check 0 element...\n")
	if table.GetNumElements() != 0 {
		t.Fatalf("empty table returned %v for GetNumElements()\n", table.GetNumElements())
	}

	fmt.Printf("... done \n")
}

/**
func TestBucket(t *testing.T) {
	table := NewTable("t", 10, 2, 64, nil, 0)
	item := &Item{1, GetBytes("value"), 5, 5}
	ok, _ := table.Insert(item)
	if !ok {
		t.Fatalf("Failed to insert item.")
	}
	if item.Bucket(table) != 5 {
		t.Fatalf("Bucket should report expected item position.")
	}
}
**/

func TestOutOfBounds(t *testing.T) {
	table := NewTable("t", 10, 2, 64, nil, 0)

	fmt.Printf("TestOutOfBounds: Insert() out of bounds...\n")
	ok, itm := table.Insert(&Item{1, GetBytes("value1"), 100, 100})
	if ok {
		t.Fatalf("Insert returned true with out of bound buckets\n")
	}
	if itm != nil {
		t.Fatalf("Insert returned wrong values\n")
	}

	fmt.Printf("TestOutOfBounds: Contains() out of bounds...\n")
	if table.Contains(&Item{1, GetBytes("value1"), 100, 100}) {
		t.Fatalf("Contains returned true with out of bound buckets\n")
	}

	fmt.Printf("TestOutOfBounds: Remove() out of bounds...\n")
	if table.Remove(&Item{1, GetBytes("value1"), 100, 100}) {
		t.Fatalf("Remove() returned true with out of bound buckets\n")
	}

	fmt.Printf("... done \n")
}

func TestFullTable(t *testing.T) {
	numBuckets := 100
	depth := 4

	capacity := numBuckets * depth
	entries := make([]Item, 0, capacity)
	table := NewTable("t", numBuckets, depth, 64, nil, 0)
	ok := true
	count := 0
	var evic *Item
	var b1, b2 int

	// Insert random values until we've reached a limit
	fmt.Printf("TestFullTable: Insert until reach capacity...\n")
	for ok {
		b1 = randBucket(numBuckets)
		b2 = randBucket(numBuckets)
		id := rand.Int()
		val := GetBytes(strconv.Itoa(rand.Int()))
		entries = append(entries, Item{id, nil, b1, b2})
		ok, evic = table.Insert(&Item{id, val, b1, b2})

		if ok {
			count++
			if !table.Contains(&Item{id, nil, b1, b2}) {
				t.Fatalf("Insert() succeeded, but Contains failed\n")
			}
			if count != table.GetNumElements() {
				t.Fatalf("Number of successful Inserts(), %v, does not match GetNumElements(), %v \n",
					count,
					table.GetNumElements())
			}
		}
	}

	// Middle count check
	fmt.Printf("TestFullTable: Fully Loaded check...\n")
	if count != table.GetNumElements() {
		t.Fatalf("Number of successful Inserts(), %v, does not match GetNumElements(), %v \n",
			count,
			table.GetNumElements())
	}
	maxCount := count

	// Remove elements one by one
	fmt.Printf("TestFullTable: Remove each element...\n")
	for _, entry := range entries {
		if !entry.Equals(evic) {
			ok = table.Remove(&entry)
			if !ok {
				t.Fatalf("Cannot Remove() an element believed to be in the table. item %d of %d",
					count,
					maxCount)
			} else {
				count--
				if count != table.GetNumElements() {
					t.Fatalf("GetNumElements()=%v returned a value that didn't match what was expected=%v \n",
						table.GetNumElements(),
						count)
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
	table := NewTable("t", 10, 2, 64, nil, 0)

	ok, itm := table.Insert(&Item{1, GetBytes("v"), 0, 1})
	if itm != nil || !ok {
		t.Fatalf("Error inserting value \n")
	}

	ok, itm = table.Insert(&Item{2, GetBytes("v"), 0, 1})
	if itm != nil || !ok {
		t.Fatalf("Error inserting value again \n")
	}

	ok, itm = table.Insert(&Item{3, GetBytes("v"), 1, 2})
	if itm != nil || !ok {
		t.Fatalf("Error inserting value in shifted buckets\n")
	}

	if !table.Remove(&Item{1, GetBytes("v"), 0, 1}) {
		t.Fatalf("Error removing value 1st time\n")
	}

	if !table.Remove(&Item{2, GetBytes("v"), 0, 1}) {
		t.Fatalf("Error removing value 2nd time\n")
	}

	if !table.Remove(&Item{3, GetBytes("v"), 1, 2}) {
		t.Fatalf("Error removing value 3rd time\n")
	}

	fmt.Printf("... done\n")
}

func TestLoadFactor(t *testing.T) {
	fmt.Printf("TestLoadFactor: ...\n")
	numBuckets := 1000
	var table *Table

	for depth := 1; depth < 32; depth *= 2 {
		count := -1
		ok := true
		table = NewTable("t", numBuckets, depth, 64, nil, int64(depth))
		for ok {
			count++
			val := GetBytes(strconv.Itoa(rand.Int()))
			ok, _ = table.Insert(&Item{rand.Int(), val, randBucket(numBuckets), randBucket(numBuckets)})
		}

		if table.GetNumElements() != count {
			t.Fatalf("Mismatch count=%v with GetNumElements()=%v \n", count, table.GetNumElements())
		}
		fmt.Printf("count=%v, depth=%v, capacity=%v, loadfactor=%v \n",
			count,
			depth,
			table.GetCapacity(),
			(float64(count) / float64(table.GetCapacity())))
	}

	fmt.Printf("... done\n")
}
