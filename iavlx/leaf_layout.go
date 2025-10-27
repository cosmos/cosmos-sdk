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
	SizeLeaf = 48
)

type LeafLayout struct {
	Id            NodeID
	KeyOffset     uint32
	OrphanVersion uint32 // TODO 5 bytes?
	Hash          [32]byte
}

func (l LeafLayout) ID() NodeID {
	return l.Id
}
