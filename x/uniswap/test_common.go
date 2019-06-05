package uniswap

import (
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	pk         = ed25519.GenPrivKey().PubKey()
	addr       = sdk.AccAddress(pk.Address())
	denom      = "atom"
	emptyAddr  sdk.AccAddress
	emptyDenom string
)
