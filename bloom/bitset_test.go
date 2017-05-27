package bloom

import (
	"fmt"
	"math"
	"testing"
)

func TestEmptyBitSet(t *testing.T) {
	b := NewBitSet(0)
	if b.Length() != 0 {
		t.Errorf("Empty set should have capacity 0, not %d", b.Length())
	}
}

func TestZeroValueBitSet(t *testing.T) {
	var b BitSet
	if b.Length() != 0 {
		t.Errorf("Empty set should have capacity 0, not %d", b.Length())
	}
}

func TestTestTooLong(t *testing.T) {
	b := new(BitSet)
	if b.Test(1) {
		t.Error("Unexpected value: true")
		return
	}
}

func TestNewPanic(t *testing.T) {
	b := NewBitSet(Cap())
	if b != nil {
		t.Error("Unexpected value: %v", b)
	}
}

func TestBitSetNew(t *testing.T) {
	b := NewBitSet(16)
	if b.Test(0) {
		t.Errorf("Unable to make a bit set and read its 0th value.")
	}
}

func TestBitSetHuge(t *testing.T) {
	b := NewBitSet(math.MaxUint32)
	if b.Test(0) {
		t.Errorf("Unable to make a huge bit set and read its 0th balue.")
	}
}

func TestLength(t *testing.T) {
	b := NewBitSet(1000)
	if b.Length() != 1000 {
		t.Errorf("Len should be 1000, but is %d.", b.Length())
	}
}

func TestBitSetIsClear(t *testing.T) {
	b := NewBitSet(1000)
	for i := uint64(0); i < 1000; i++ {
		if b.Test(i) {
			t.Errorf("Bit %d is set, and it shouldn't be.", i)
		}
	}
}

func TestBitSetAndGet(t *testing.T) {
	b := NewBitSet(1000)
	b.Set(100)
	if !b.Test(100) {
		t.Errorf("Bit %d is clear, and it shouldn't be.", 100)
	}
}

func TestSetTo(t *testing.T) {
	b := NewBitSet(1000)
	b.SetTo(100, true)
	if !b.Test(100) {
		t.Errorf("Bit %d is clear, and it shouldn't be.", 100)
	}
	b.SetTo(100, false)
	if b.Test(100) {
		t.Errorf("Bit %d is set, and it shouldn't be.", 100)
	}
}

func TestOutOfBoundsLong(t *testing.T) {
	b := NewBitSet(64)
	defer func() {
		if r := recover(); r != nil {
			t.Error("Long distance out of index error should not have caused a panic")
		}
	}()
	b.Set(1000)
}

func TestOutOfBoundsClose(t *testing.T) {
	b := NewBitSet(65)
	defer func() {
		if r := recover(); r != nil {
			t.Error("Local out of index error should not have caused a panic")
		}
	}()
	b.Set(66)
}

// nil tests
func TestNullTest(t *testing.T) {
	var b *BitSet
	defer func() {
		if r := recover(); r == nil {
			t.Error("Checking bit of null reference should have caused a panic")
		}
	}()
	b.Test(66)
}

func TestNullSet(t *testing.T) {
	var b *BitSet
	defer func() {
		if r := recover(); r == nil {
			t.Error("Setting bit of null reference should have caused a panic")
		}
	}()
	b.Set(66)
}

func TestNullClear(t *testing.T) {
	var b *BitSet
	defer func() {
		if r := recover(); r == nil {
			t.Error("Clearning bit of null reference should have caused a panic")
		}
	}()
	b.Clear(66)
}

func TestEqual(t *testing.T) {
	a := NewBitSet(100)
	b := NewBitSet(99)
	c := NewBitSet(100)
	if Equal(a, b) {
		t.Error("Sets of different sizes should be not be equal")
	}
	if !Equal(a, c) {
		t.Error("Two empty sets of the same size should be equal")
	}
	a.Set(99)
	c.Set(0)
	if Equal(a, c) {
		t.Error("Two sets with differences should not be equal")
	}
	c.Set(99)
	a.Set(0)
	if !Equal(a, c) {
		t.Error("Two sets with the same bits set should be equal")
	}
	if Equal(a, nil) {
		t.Error("The sets should be different")
	}
	a = NewBitSet(0)
	b = NewBitSet(0)
	if !Equal(a, b) {
		t.Error("Two empty set should be equal")
	}
}

func TestFrom(t *testing.T) {
	u := []uint64{2, 3, 5, 7, 11}
	b := From(uint64(len(u)*8), u)
	outType := fmt.Sprintf("%T", b)
	expType := "*bloom.BitSet"
	if outType != expType {
		t.Error("Expecting type: ", expType, ", gotf:", outType)
		return
	}
}

func TestBytes(t *testing.T) {
	b := new(BitSet)
	c := b.Bytes()
	outType := fmt.Sprintf("%T", c)
	expType := "[]uint64"
	if outType != expType {
		t.Error("Expecting type: ", expType, ", gotf:", outType)
		return
	}
	if len(c) != 0 {
		t.Error("The slice should be empty")
		return
	}
}
