package cuckootable

import (
	"log"
)

type Entry struct {
	Bucket1 uint32
	Bucket2 uint32
	Data    interface{}
}

type Bucket struct {
	entries []*Entry
	filled  []bool
}

type Table struct {
	numBuckets int
	depth      int
	buckets    []*Bucket
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

// Inserts the entry into the cuckoo table
func (t *Table) Insert(e *Entry) {
}

// Tries to inserts `target` into specified bucket
// bucketIndex must be either `target.Bucket1` or `target.Bucket2` or nothing happens
// If the bucket is already full, skip
// Returns true if success, false if bucket already full
func (t *Table) tryInsertToBucket(bucketIndex uint32, target *Entry) bool {
	if target.Bucket1 != bucketIndex && target.Bucket2 != bucketIndex {
		return false
	}

	for i, b := range t.buckets[bucketIndex].filled {
	}
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
