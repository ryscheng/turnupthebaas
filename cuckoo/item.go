package cuckoo

// Item holds a full item data for cuckoo table placement.
type Item struct {
	ID      int
	Data    []byte
	Bucket1 int
	Bucket2 int
}

// Copy duplicates an Item.
func (i Item) Copy() *Item {
	other := &Item{}
	other.ID = i.ID
	other.Data = make([]byte, len(i.Data))
	other.Bucket1 = i.Bucket1
	other.Bucket2 = i.Bucket2
	copy(other.Data, i.Data)
	return other
}

// Equals compares item equality.
func (i *Item) Equals(other *Item) bool {
	if other == nil {
		return false
	}
	return i.Bucket1 == other.Bucket1 &&
		i.Bucket2 == other.Bucket2 &&
		i.ID == other.ID
}
