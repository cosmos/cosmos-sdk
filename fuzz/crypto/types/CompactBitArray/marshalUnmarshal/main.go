package marshalunmarshal

import (
	"github.com/cosmos/cosmos-sdk/crypto/types"
)

func Fuzz(data []byte) int {
	cba, err := types.CompactUnmarshal(data)
	if err != nil {
		return 0
	}
	if cba == nil {
		panic("Inconsistency, no error, yet BitArray is nil")
	}
	return 1
}
