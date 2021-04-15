package addressfrombalancesstore

import (
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func Fuzz(data []byte) int {
	_, err := types.AddressFromBalancesStore(data)
	if err != nil {
		return 1
	}
	return 0
}
