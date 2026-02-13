package tree

import "bytes"

type KeyItem interface {
	GetKey() []byte
}

func KeyItemLess[T KeyItem](a, b T) bool {
	return bytes.Compare(a.GetKey(), b.GetKey()) < 0
}
