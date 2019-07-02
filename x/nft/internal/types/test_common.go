package types

import (
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// nolint:deadcode unused
var (
	userPk1   = ed25519.GenPrivKey().PubKey()
	userPk2   = ed25519.GenPrivKey().PubKey()
	userAddr1 = sdk.AccAddress(userPk1.Address())
	userAddr2 = sdk.AccAddress(userPk2.Address())
)
