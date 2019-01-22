package types

import (
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	delPk1       = ed25519.GenPrivKey().PubKey()
	delPk2       = ed25519.GenPrivKey().PubKey()
	delPk3       = ed25519.GenPrivKey().PubKey()
	delAddr1     = sdk.AccAddress(delPk1.Address())
	delAddr2     = sdk.AccAddress(delPk2.Address())
	delAddr3     = sdk.AccAddress(delPk3.Address())
	emptyDelAddr sdk.AccAddress

	valPk1       = ed25519.GenPrivKey().PubKey()
	valPk2       = ed25519.GenPrivKey().PubKey()
	valPk3       = ed25519.GenPrivKey().PubKey()
	valAddr1     = sdk.ValAddress(valPk1.Address())
	valAddr2     = sdk.ValAddress(valPk2.Address())
	valAddr3     = sdk.ValAddress(valPk3.Address())
	emptyValAddr sdk.ValAddress

	emptyPubkey crypto.PubKey
)
