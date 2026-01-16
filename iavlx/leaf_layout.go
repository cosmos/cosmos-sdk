package iavlx

import (
	"fmt"
	"unsafe"
)

func init() {
	if unsafe.Sizeof(LeafLayout{}) != SizeLeaf {
		panic(fmt.Sprintf("invalid LeafLayout size: got %d, want %d", unsafe.Sizeof(LeafLayout{}), SizeLeaf))
	}
}

const (
	SizeLeaf = 44
)

type LeafLayout struct {
	Id        NodeID
	Hash      [32]byte
	KeyOffset uint32 // TODO check if we have extra padding here
}

func (l LeafLayout) ID() NodeID {
	return l.Id
}
