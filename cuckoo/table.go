package cuckoo

import (
	"math/rand"

	"github.com/privacylab/talek/common"
)

// MaxEvictions represents the number of chained evictions before an insert
// attempt fails.
const MaxEvictions int = 500

// ItemLocation holds metadata for item placement in a cuckoo table.
type ItemLocation struct {
	id      int
	filled  bool
	bucket1 int
	bucket2 int
}

// Table is a cuckoo table managing placement of Items.
type Table struct {
	name        string
	numBuckets  int    // Number of buckets
	bucketDepth int    // Items in each bucket
	itemSize    int    // number of bytes in an item
	data        []byte //
	rand        *rand.Rand
	log         *common.Logger
	index       []ItemLocation
}

// NewTable creates a new cuckoo table optionaly backed by a pre-allocated memory area.
// Two cuckoo tables will have identical state iff,
// 1. the same randSeed is used
// 2. the same operations are applied in the same order
// numBuckets = number of buckets
// depth = the number of entries per bucket
// randSeed = seed for PRNG
func NewTable(name string, numBuckets int, bucketDepth int, itemSize int,
	data []byte, randSeed int64) *Table {
	t := &Table{name, numBuckets, bucketDepth, itemSize, nil, nil, nil, nil}
	if data == nil {
		data = make([]byte, numBuckets*bucketDepth*itemSize)
	}
	t.data = data
	t.rand = rand.New(rand.NewSource(randSeed))
	t.log = common.NewLogger(name)
	t.index = make([]ItemLocation, numBuckets*bucketDepth)

	if len(data) != numBuckets*bucketDepth*itemSize {
		t.log.Error.Printf("NewTable(%v) failed: len(data)=%v is not equal to numBuckets*bucketDepth*itemSize (%v,%v,%v)", name, len(data), numBuckets, bucketDepth, itemSize)
		return nil
	}

	return t
}

/********************
 * PUBLIC METHODS
 ********************/

// GetCapacity returns the total capacity of the table (numBuckets * depth)
func (t *Table) GetCapacity() int {
	return t.numBuckets * t.bucketDepth
}

// GetNumElements returns the number of elements stored in the table
// Load factor = GetNumElements() / GetCapacity()
func (t *Table) GetNumElements() int {
	result := 0
	for _, itemLocation := range t.index {
		if itemLocation.filled {
			result++
		}
	}

	return result
}

// Bucket returns the bucket in a table that the Item is in, if it is in the table.
// -1 otherwise
func (t *Table) Bucket(item *Item) int {
	if t.isInBucket(item.Bucket1, item) {
		return item.Bucket1
	} else if t.isInBucket(item.Bucket2, item) {
		return item.Bucket2
	}
	return -1
}

// Contains checks if value exists in specified buckets
// the value must have been inserted with the same bucket1 and bucket2 values
// Returns:
// - true if the item is in either bucket
// - false if either bucket is out of range
// - false if value not in either bucket
func (t *Table) Contains(item *Item) bool {
	if item.Bucket1 >= t.numBuckets || item.Bucket2 >= t.numBuckets {
		return false
	}

	return t.isInBucket(item.Bucket1, item) || t.isInBucket(item.Bucket2, item)
}

// Insert adds item into the cuckoo table, even if a duplicate value already
// exists in table. Returns:
// - true on success, false on failure
// - false if item.Data is not equal to t.itemSize
// - false if either bucket is out of range
// - false if insertion cannot complete because reached MAX_EVICTIONS
func (t *Table) Insert(item *Item) (bool, *Item) {
	var nextBucket int
	if item.Bucket1 >= t.numBuckets || item.Bucket2 >= t.numBuckets {
		return false, nil
	}

	// Check item data size
	if len(item.Data) != t.itemSize {
		return false, nil
	}

	// Randomly select 1 bucket first
	coin := t.rand.Int() % 2 // Coin can be 0 or 1
	if coin == 0 {
		if t.tryInsertToBucket(item.Bucket1, item) {
			return true, nil
		}
		nextBucket = item.Bucket2
	} else {
		if t.tryInsertToBucket(item.Bucket2, item) {
			return true, nil
		}
		nextBucket = item.Bucket1
	}

	// Then try the other bucket, starting the eviction loop
	var ok bool
	for i := 0; i < MaxEvictions; i++ {
		if ok, item = t.insertAndEvict(nextBucket, item); !ok {
			t.log.Error.Fatalf("Lost item. Evicted, but was unable to add.")
			return false, item
		} else if item == nil {
			return true, nil
		} else if item.Bucket1 == nextBucket {
			nextBucket = item.Bucket2
		} else {
			nextBucket = item.Bucket1
		}
	}

	//t.log.Printf("Insert: MAX %v evictions\n", MAX_EVICTIONS)
	return false, item
}

// Remove deletes item from the cuckoo table, looking in only 2 specified buckets
// Only matches if the value was previously inserted with the same {bucket1, bucket2} values
// If the incorrect buckets were specified, it won't go searching for you
// If the value exists in the table multiple times, it will only remove one
// Returns:
// - true if a value was removed from either bucket, false if not
// - fails if either bucket is out of range
func (t *Table) Remove(item *Item) bool {
	if item.Bucket1 >= t.numBuckets || item.Bucket2 >= t.numBuckets {
		//t.log.Error.Fatalf("Failed to remove item with invalid buckets.")
		return false
	}

	var result bool
	var nextBucket int
	coin := t.rand.Int() % 2 // Coin can be 0 or 1
	if coin == 0 {
		result = t.removeFromBucket(item.Bucket1, item)
		nextBucket = item.Bucket2
	} else {
		result = t.removeFromBucket(item.Bucket2, item)
		nextBucket = item.Bucket1
	}

	return result || t.removeFromBucket(nextBucket, item)
}

/********************
 * PRIVATE METHODS
 ********************/

// Checks if the `value` is in a specified bucket
// - bucket MUST be within bounds
// Returns: the true if `value.Equals(...)`
// returns true for any value in bucket, false if not present
func (t *Table) isInBucket(bucketIndex int, item *Item) bool {
	for i := 0; i < t.bucketDepth; i++ {
		idx := t.bucketDepth*bucketIndex + i
		if t.index[idx].filled &&
			t.index[idx].bucket1 == item.Bucket1 &&
			t.index[idx].bucket2 == item.Bucket2 &&
			t.index[idx].id == item.ID {
			return true
		}
	}
	return false
}

// Tries to inserts an item into specified bucket
// If the bucket is already full, no-op
// Preconditions:
// - bucket MUST be within bounds
// - item MUST contain data of size t.itemSize
// Returns: true if success, false if bucket already full
func (t *Table) tryInsertToBucket(bucketIndex int, item *Item) bool {
	// Search for an empty slot
	for i := bucketIndex * t.bucketDepth; i < (bucketIndex+1)*t.bucketDepth; i++ {
		if !t.index[i].filled {
			copy(t.data[i*t.itemSize:], item.Data)
			t.index[i].id = item.ID
			t.index[i].bucket1 = item.Bucket1
			t.index[i].bucket2 = item.Bucket2
			t.index[i].filled = true
			return true
		}
	}

	return false
}

// Tries to insert `bucketLoc, value` into specified bucket
// Preconditions:
// - bucket MUST be within bounds
// - item MUST contain data of size t.itemSize
// Returns:
// - (-1, BucketLocation{}, nil, true) if there's empty space and succeeds
// - false if insertion triggered an eviction
//   other values contain the evicted item's alternate bucket, BucketLocation pair, and value
func (t *Table) insertAndEvict(bucketIndex int, item *Item) (bool, *Item) {
	if item.Bucket1 != bucketIndex && item.Bucket2 != bucketIndex {
		return false, item
	}
	if t.tryInsertToBucket(bucketIndex, item) {
		return true, nil
	}

	// Eviction
	itemIndex := bucketIndex*t.bucketDepth + (t.rand.Int() % t.bucketDepth)
	removedItem := t.getItem(itemIndex).Copy()
	t.index[itemIndex].filled = false

	if !t.tryInsertToBucket(bucketIndex, item) {
		t.log.Error.Fatalf("insertAndEvict: no space in bucket after eviction!")
		return false, removedItem
	}
	return true, removedItem
}

// Removes a single copy of `value` from the specified bucket
// bucketLoc and value must both match
// Preconditions:
// - bucket MUST be within bounds
// Returns: true if succeeds, false if value not in bucket
func (t *Table) removeFromBucket(bucketIndex int, item *Item) bool {
	for i := bucketIndex * t.bucketDepth; i < (bucketIndex+1)*t.bucketDepth; i++ {
		if item != nil && item.Equals(t.getItem(i)) {
			t.index[i].filled = false
			return true
		}
	}
	return false
}

func (t *Table) getItem(itemIndex int) *Item {
	if !t.index[itemIndex].filled {
		return nil
	}
	return &Item{
		t.index[itemIndex].id,
		t.data[itemIndex*t.itemSize : (itemIndex+1)*t.itemSize],
		t.index[itemIndex].bucket1,
		t.index[itemIndex].bucket2}
}
