package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

var (
	pk1   = crypto.GenPrivKeyEd25519().PubKey()
	pk2   = crypto.GenPrivKeyEd25519().PubKey()
	pk3   = crypto.GenPrivKeyEd25519().PubKey()
	addr1 = sdk.AccAddress(pk1.Address())
	addr2 = sdk.AccAddress(pk2.Address())
	addr3 = sdk.AccAddress(pk3.Address())

	emptyAddr   sdk.AccAddress
	emptyPubkey crypto.PubKey
)
