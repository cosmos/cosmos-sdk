package marshalunmarshal

import (
	"github.com/cosmos/cosmos-sdk/crypto/types"
)

func Fuzz(data []byte) int {
	cba, err := types.CompactUnmarshal(data)
	if err != nil {
		return 0
	}
	if cba == nil && string(data) != "null" {
		panic("Inconsistency, no error, yet BitArray is nil")
	}
	if cba.SetIndex(-1, true) {
		panic("Set negative index success")
	}
	if cba.GetIndex(-1) {
		panic("Get negative index success")
	}
	return 1
}
