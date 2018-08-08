package slashing

import (
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// InitGenesis initializes the keeper's address to pubkey map
func InitGenesis(keeper Keeper, data types.GenesisState) {
	for _, validator := range data.Validators {
		keeper.addPubkey(validator.GetPubKey(), false)
	}
	return
}
