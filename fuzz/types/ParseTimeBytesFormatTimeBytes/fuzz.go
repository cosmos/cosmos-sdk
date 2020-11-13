package parsetimebytes

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
)

func Fuzz(bin []byte) int {
	t, err := types.ParseTimeBytes(bin)
	if err != nil {
		return -1
	}
	brt := types.FormatTimeBytes(t)
	if !bytes.Equal(brt, bin) {
		panic(fmt.Sprintf("Roundtrip failure, got\n%s", brt))
	}

	// Parsed successfully, indicate to the fuzzer that it should increase
	// the priority of this input, thus make it a part of the corpus.
	return 1
}
