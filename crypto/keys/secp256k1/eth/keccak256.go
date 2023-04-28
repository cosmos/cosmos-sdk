package eth

import (
	io "io"

	"golang.org/x/crypto/sha3"
)

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func Keccak256(data ...[]byte) []byte {
	b := make([]byte, 32)
	d := sha3.NewLegacyKeccak256().(io.ReadWriter)
	for _, b := range data {
		_, _ = d.Write(b)
	}
	_, _ = d.Read(b)
	return b
}
