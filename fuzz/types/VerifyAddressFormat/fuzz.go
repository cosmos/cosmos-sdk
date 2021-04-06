package verifyaddressformat

import (
	"github.com/cosmos/cosmos-sdk/types"
)

func Fuzz(data []byte) int {
	if types.VerifyAddressFormat(data) != nil {
		return 0
	}
	return 1
}
