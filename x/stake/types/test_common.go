package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

var (
	// dummy pubkeys/addresses
	pk1   = crypto.GenPrivKeyEd25519().PubKey()
	pk2   = crypto.GenPrivKeyEd25519().PubKey()
	addr1 = pk1.Address()
	addr2 = pk2.Address()

	emptyAddr   sdk.Address
	emptyPubkey crypto.PubKey
)
