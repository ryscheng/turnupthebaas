package pir

type BitVec struct {
	Bits int
	Data []uint64
}

func NewBitVec(bits int) *BitVec {
	v := new(BitVec)
	v.Resize(bits)
	return v
}

func (v *BitVec) Len() int {
	return v.Bits
}

func (v *BitVec) Resize(bits int) {
	if bits < 0 {
		panic("negative size")
	}
	count := bits / 64
	if bits%64 != 0 {
		count++
	}
	data := make([]uint64, count)
	copy(data, v.Data)
	v.Bits = bits
	v.Data = data
}

func (v *BitVec) IsSet(n int) bool {
	if n < 0 || n >= v.Len() {
		panic("bad index")
	}
	return (v.Data[n/64]&(1<<(uint(n)%64)) != 0)
}

func (v *BitVec) Set(n int) {
	v.Data[n/64] |= 1 << (uint(n) % 64)
}

func (v *BitVec) Clear(n int) {
	v.Data[n/64] &^= 1 << (uint(n) % 64)
}
