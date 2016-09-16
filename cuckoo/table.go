package cuckoo

import (
	"log"
	"math/rand"
	"os"
)

type Comparable interface {
	Equals(other Comparable) bool
}

type BucketLocation struct {
	Bucket1 int
	Bucket2 int
}

type Bucket struct {
	// `entries` and `filled` must be the same size
	data      []*Comparable     // Stores actual data. Validity of an entry determined by `filled`
	bucketLoc []*BucketLocation // Stores the 2 bucket locations for each entry
	filled    []bool            // False if cell is empty. Only read `t.entries[i]` if `t.filled[i]==true`
}

type Table struct {
	log        *log.Logger
	name       string
	numBuckets int // Number of buckets
	depth      int // Capacity of each bucket
	rand       *rand.Rand

	buckets []*Bucket // Data
}

// Creates a brand new cuckoo table
// Two cuckoo tables will have identical state iff,
// 1. the same randSeed is used
// 2. the same operations are applied in the same order
// numBuckets = number of buckets
// depth = the number of entries per bucket
// randSeed = seed for PRNG
func NewTable(name string, numBuckets int, depth int, randSeed int64) *Table {
	t := &Table{}
	t.log = log.New(os.Stdout, "[Cuckoo:"+name+"] ", log.Ldate|log.Ltime|log.Lshortfile)
	t.name = name
	t.numBuckets = numBuckets
	t.depth = depth
	t.rand = rand.New(rand.NewSource(randSeed))

	t.buckets = make([]*Bucket, numBuckets)
	for i := 0; i < numBuckets; i++ {
		t.buckets[i] = &Bucket{}
		t.buckets[i].data = make([]*Comparable, depth)
		t.buckets[i].bucketLoc = make([]*BucketLocation, depth)
		// We assume this will be filled with `false` as per bool's default value
		t.buckets[i].filled = make([]bool, depth)
	}
	return t
}

/********************
 * PUBLIC METHODS
 ********************/

/**
// Checks if entry exists in the table
// Returns true if an entry exists where all fields match
func (t *Table) Contains(bucket1 int, bucket2, value *Comparable) bool {
	result := false
	if e.Bucket1 < t.numBuckets {
		result = result || t.isInBucket(e.Bucket1, e)
	}
	if e.Bucket2 < t.numBuckets {
		result = result || t.isInBucket(e.Bucket2, e)
	}
	return result
}

// Inserts the entry into the cuckoo table
// Returns true on success, false if not inserted
// Even if false is returned, the underlying data structure might be different (e.g. rebuilt)
func (t *Table) Insert(bucket1, bucket2, value *Comparable) bool {
	coin := rand.Int31()
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

// Removes the entry from the cuckoo table
func (t *Table) Remove(bucket1 int, bucket2 int, target *Comparable) {
	t.removeFromBucket(target.Bucket1, target)
	t.removeFromBucket(target.Bucket2, target)
}
**/

/********************
 * PRIVATE METHODS
 ********************/

/**
// Checks if the `target` is in a specified bucket
// Returns true if an entry exists where all fields match
func (t *Table) isInBucket(bucketIndex int, target *Entry) bool {
	bucket := t.buckets[bucketIndex]
	for i := 0; i < t.depth; i++ {
		if bucket.filled[i] && bucket.entries[i].Equals(target) {
			return true
		}
	}
	return false
}

// Tries to inserts `target` into specified bucket
// bucketIndex must be either `target.Bucket1` or `target.Bucket2` or nothing happens
// If the bucket is already full, skip
// Returns true if success, false if bucket already full
func (t *Table) tryInsertToBucket(bucketIndex int, target *Entry) bool {
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

func (t *Table) evictAndInsert(bucketIndex int, target *Entry) *Entry {
}

// Removes all copies of `target` from the specified bucket
// `target` matches against any entry where all fields match
func (t *Table) removeFromBucket(bucketIndex int, target *Entry) bool {

}
**/
