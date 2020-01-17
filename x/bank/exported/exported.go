package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisBalance defines a genesis balance interface that holds an address and balance
type GenesisBalance interface {
	GetAddress() sdk.AccAddress
	GetCoins() sdk.Coins
}
