package cuckootable

import (
	"math/rand"
)

type Entry struct {
	Bucket1 uint32
	Bucket2 uint32
	Data    interface{}
}

type Bucket struct {
	// `entries` and `filled` must be the same size
	entries []*Entry //Stores actual entries. Validity of an entry determined by `filled`
	filled  []bool   //False if cell is empty. Only read `t.entries[i]` if `t.filled[i]==true`
}

type Table struct {
	numBuckets int       // Number of buckets
	depth      int       // Capacity of each bucket
	buckets    []*Bucket // Data
}

// Creates a brand new cuckoo table
// numBuckets = number of buckets
// depth = the number of entries per bucket
func NewTable(numBuckets int, depth int) *Table {
	t := &Table{}
	t.numBuckets = numBuckets
	t.depth = depth
	t.buckets = make([]*Bucket, numBuckets)
	for i := 0; i < numBuckets; i++ {
		t.buckets[i] = &Bucket{}
		t.buckets[i].entries = make([]*Entry, depth)
		// We assume this will be filled with `false` as per bool's default value
		t.buckets[i].filled = make([]bool, depth)
	}
	return t
}

// Checks if entry exists in the table
// Returns true if an entry exists where all fields match
func (t *Table) Contains(e *Entry) bool {

}

// Inserts the entry into the cuckoo table
// Returns true on success, false if not inserted
// Even if false is returned, the underlying data structure might be different (e.g. rebuilt)
func (t *Table) Insert(e *Entry) bool {
	ok := t.tryInsertToBucket(e.Bucket1, e)
	if ok {
		return true
	}
	ok = t.tryInsertToBucket(e.Bucket2, e)
	if ok {
		return true
	}
	// @todo Evict

}

// Tries to inserts `target` into specified bucket
// bucketIndex must be either `target.Bucket1` or `target.Bucket2` or nothing happens
// If the bucket is already full, skip
// Returns true if success, false if bucket already full
func (t *Table) tryInsertToBucket(bucketIndex uint32, target *Entry) bool {
	// Assert bucketIndex is part of `target`
	if target.Bucket1 != bucketIndex && target.Bucket2 != bucketIndex {
		return false
	}

	// Search for an empty slot
	bucket := t.buckets[bucketIndex]
	for i, filled := range bucket.filled {
		if !filled {
			bucket.filled[i] = true
			bucket.entries[i] = target
			return true
		}
	}

	return false
}

func (t *Table) evictAndInsert(bucketIndex uint32, target *Entry) *Entry {
}

// Removes the entry from the cuckoo table
func (t *Table) Remove(target *Entry) {
	t.removeFromBucket(target.Bucket1, target)
	t.removeFromBucket(target.Bucket2, target)
}

// Removes all copies of `target` from the specified bucket
// `target` matches against any entry where all fields match
func (t *Table) removeFromBucket(bucketIndex uint32, target *Entry) bool {

}
