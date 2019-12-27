package types

func NewBitArray(hint int) BitArray {
	quo := hint / 8
	rem := hint % 8
	if rem != 0 {
		quo++
	}
	return make(BitArray, quo)
}

type BitArray []byte

func (b BitArray) position(index int) int {
	return index / 8
}

func (b BitArray) bit(index int) int {
	return index % 8
}

func (b BitArray) Clear(index int) {
	b.Set(index, 0)
}

func (b BitArray) Fill(index int) {
	b.Set(index, 1)
}

func (b BitArray) Set(index, val int) {
	pos := b.position(index)
	if pos >= len(b) {
		panic("index overflow")
	}
	bit := b.bit(index)
	if val != 0 {
		b[pos] |= 1 << bit
	} else {
		b[pos] &= ^(1 << bit)
	}
}

func (b BitArray) Get(index int) bool {
	pos := b.position(index)
	if pos >= len(b) {
		panic("index overlow")
	}
	return (b[pos] & (1 << b.bit(index))) > 0
}
