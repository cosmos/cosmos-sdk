package tree

import (
	"crypto/sha256"
	"testing"
)

func emptyHashF() []byte {
	h := sha256.Sum256([]byte{})
	return h[:]
}

func TestEmptyHashEqual(t *testing.T) {
	var emptyHashBytes [32]byte
	copy(emptyHashBytes[:], emptyHashF())
	expected := emptyHash
	if emptyHashBytes != expected {
		t.Fatalf("empty hash mismatch: got=%#v expected=%#v", emptyHashBytes, expected)
	}
}
