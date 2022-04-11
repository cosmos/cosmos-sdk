package addressfrombalancesstore

import (
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func Fuzz(data []byte) int {
	_, _, err := types.AddressAndDenomFromBalancesStore(data)
	if err != nil {
		return 1
	}
	return 0
}
