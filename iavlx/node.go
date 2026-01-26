package iavlx

import "fmt"

type Node interface {
	ID() NodeID
	Height() uint8
	IsLeaf() bool
	Size() int64
	Version() uint32
	Key() ([]byte, error)
	Value() ([]byte, error)
	Left() *NodePointer
	Right() *NodePointer
	Hash() []byte
	SafeHash() []byte
	MutateBranch(version uint32) (*MemNode, error)
	Get(key []byte) (value []byte, index int64, err error)

	fmt.Stringer
}
