package server

// Table represents the logic of maintaining and updating writes into a cuckoo
// table for the PDB server. It maintains two regions of memory, an 'active'
// read-only region that reads occur on, and an 'alternate' region that writes
// are made to. The two regions can flip-flop to update to the next next
// snapshot for reads.

import (
	"github.com/ryscheng/pdb/cuckoo"
	"github.com/ryscheng/pdb/pir"
	"log"
)

type Table struct {
	log *log.Logger

	server *pir.PirServer

	activeDB      *pir.PirDB
	activeTable   *cuckoo.Table
	activeEntries []cuckoo.Item

	alternateDB      *pir.PirDB
	alternateTable   *cuckoo.Table
	alternateEntries []cuckoo.Item

	pendingWrites []cuckoo.Item

	numBuckets  int
	bucketDepth int
	// at what load factor should the table pro-actively remove items?
	maxLoadFactor float32
	// when removing items, what fraction should be removed?
	loadFactorStep float32
}

func NewTable(server *pir.PirServer, name string, log *log.Logger, depth int, maxLoadFactor float32, loadFactorStep float32) *Table {
	table := &Table{}
	table.log = log
	table.server = server
	table.numBuckets = server.CellCount
	table.bucketDepth = depth
	table.maxLoadFactor = maxLoadFactor
	table.loadFactorStep = loadFactorStep

	activeDB, err := server.GetDB()
	if err != nil {
		table.log.Fatalf("Could not allocate DB region: %v", err)
		return nil
	}
	table.activeDB = activeDB
	//TODO: randseed. set from somewhere.
	table.activeTable = cuckoo.NewTable(name+"-A", table.numBuckets, table.bucketDepth, server.CellLength/table.bucketDepth, table.activeDB.DB, 0)
	table.activeEntries = make([]cuckoo.Item, 0, table.numBuckets*table.bucketDepth)

	alternateDB, err := server.GetDB()
	if err != nil {
		table.log.Fatalf("Could not allocate DB region: %v", err)
		return nil
	}
	table.alternateDB = alternateDB
	table.alternateTable = cuckoo.NewTable(name+"-B", table.numBuckets, table.bucketDepth, server.CellLength/table.bucketDepth, table.alternateDB.DB, 0)
	table.alternateEntries = make([]cuckoo.Item, 0, table.numBuckets*table.bucketDepth)

	table.pendingWrites = make([]cuckoo.Item, 0, table.numBuckets*table.bucketDepth)

	return table
}

func (t *Table) Flop() error {
	newAlternateDB := t.activeDB
	newAlternateTable := t.activeTable
	newAlternateEntries := t.activeEntries

	err := t.server.SetDB(t.alternateDB)
	if err != nil {
		return err
	}

	t.activeDB = t.alternateDB
	t.activeTable = t.alternateTable
	t.activeEntries = t.alternateEntries

	t.alternateDB = newAlternateDB
	t.alternateTable = newAlternateTable
	t.alternateEntries = newAlternateEntries

	// alternate is now out of date. re-apply 'pendingWrites'.
	repend := t.pendingWrites
	t.pendingWrites = make([]cuckoo.Item, 0, len(repend))
	for item := range repend {
		t.Write(&repend[item])
	}
	t.pendingWrites = repend[0:0]
	return nil
}

func (t *Table) Write(item *cuckoo.Item) error {
	// Mark this item as one that needs to go in the active table when swapped out.
	t.pendingWrites = append(t.pendingWrites, *item)

	// Apply immediately to alternate table.
	t.alternateEntries = append(t.alternateEntries, *item)
	ok, evicted := t.alternateTable.Insert(item)
	if !ok || len(t.alternateEntries) > int(float32(t.numBuckets*t.bucketDepth)*t.maxLoadFactor) {
		toRemove := int(float32(t.numBuckets*t.bucketDepth) * t.loadFactorStep)
		for i := 0; i < toRemove; i++ {
			t.alternateTable.Remove(&t.alternateEntries[i])
		}
		t.alternateEntries = t.alternateEntries[toRemove:]
		// Trigger eviction.
		if evicted != nil {
			return t.Write(evicted)
		}
	}
	return nil
}

func (t *Table) Close() {
	t.activeDB.Free()
	t.alternateDB.Free()
}
