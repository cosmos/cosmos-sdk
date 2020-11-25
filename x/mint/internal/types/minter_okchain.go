package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MinterCustom struct {
	NextBlockToUpdate uint64 `json:"next_block_to_update" yaml:"next_block_to_update"` // record the block height for next year
	MintedPerBlock types.DecCoins `json:"minted_per_block" yaml:"minted_per_block"` // record the MintedPerBlock per block in this year
}

// NewMinterCustom returns a new Minter object with the given inflation and annual
// provisions values.
func NewMinterCustom(nextBlockToUpdate uint64, mintedPerBlock sdk.DecCoins) MinterCustom {
	return MinterCustom{
		NextBlockToUpdate: nextBlockToUpdate,
		MintedPerBlock: mintedPerBlock,
	}
}

// InitialMinterCustom returns an initial Minter object with a given inflation value.
func InitialMinterCustom() MinterCustom {
	return NewMinterCustom(
		0,
		sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.ZeroInt())},
	)
}

// DefaultInitialMinterCustom returns a default initial MinterCustom object for a new chain
// which uses an inflation rate of 1%.
func DefaultInitialMinterCustom() MinterCustom {
	return InitialMinterCustom()
}

// ValidateMinterCustom validate minter
func ValidateMinterCustom(minter MinterCustom) error {
	if len(minter.MintedPerBlock) != 1 || minter.MintedPerBlock[0].Denom != sdk.DefaultBondDenom {
		return fmt.Errorf(" MintedPerBlock must contain only %s", sdk.DefaultBondDenom)
	}
	return nil
}
