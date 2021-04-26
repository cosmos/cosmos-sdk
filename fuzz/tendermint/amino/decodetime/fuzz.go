package decodetime

import (
	"fmt"

	amino "github.com/tendermint/go-amino"
)

func Fuzz(data []byte) int {
	if len(data) == 0 {
		return -1
	}
	t, n, err := amino.DecodeTime(data)
	if err != nil {
		return -1
	}
	if n < 0 {
		panic(fmt.Sprintf("n=%d < 0", n))
	}
	if t.IsZero() {
		return 0
	}
	return 1
}
