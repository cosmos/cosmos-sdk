package slashing

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

// InitGenesis initializes the keeper's address to pubkey map
func InitGenesis(keeper Keeper, data types.GenesisState) {
	for _, validator := range data.Validators {
		pubkey := validator.GetPubKey()
		addr := new([tmhash.Size]byte)
		copy(addr[:], pubkey.Bytes())
		keeper.addressToPubkey[*addr] = pubkey
		fmt.Println("WAS THIS CALLED?!?!??!")
		fmt.Println(keeper.addressToPubkey)
	}
	return
}
