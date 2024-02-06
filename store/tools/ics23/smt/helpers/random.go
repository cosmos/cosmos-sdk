package helpers

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	rand "math/rand"
)

const (
	strChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func init() {
	rand.Seed(42)
	// rand.Seed(crandSeed())
}

func randStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = strChars[rand.Intn(len(strChars))]
	}
	return string(b)
}

func crandSeed() int64 {
	var seed int64
	err := binary.Read(crand.Reader, binary.BigEndian, &seed)
	if err != nil {
		panic(fmt.Sprintf("could not read random seed from crypto/rand: %v", err))
	}
	return seed
}
