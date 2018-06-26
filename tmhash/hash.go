package tmhash

import (
	"crypto/sha256"
	"hash"
)

const (
	Size      = 20
	BlockSize = sha256.BlockSize
)

type sha256trunc struct {
	sha256 hash.Hash
}

func (h sha256trunc) Write(p []byte) (n int, err error) {
	return h.sha256.Write(p)
}
func (h sha256trunc) Sum(b []byte) []byte {
	shasum := h.sha256.Sum(b)
	return shasum[:Size]
}

func (h sha256trunc) Reset() {
	h.sha256.Reset()
}

func (h sha256trunc) Size() int {
	return Size
}

func (h sha256trunc) BlockSize() int {
	return h.sha256.BlockSize()
}

// New returns a new hash.Hash.
func New() hash.Hash {
	return sha256trunc{
		sha256: sha256.New(),
	}
}

// Sum returns the first 20 bytes of SHA256 of the bz.
func Sum(bz []byte) []byte {
	hash := sha256.Sum256(bz)
	return hash[:Size]
}
