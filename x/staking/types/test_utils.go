package types

import (
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	pk1   = sdk.ConsPubKeyFromCryptoPubKey(ed25519.GenPrivKey().PubKey())
	pk2   = sdk.ConsPubKeyFromCryptoPubKey(ed25519.GenPrivKey().PubKey())
	pk3   = sdk.ConsPubKeyFromCryptoPubKey(ed25519.GenPrivKey().PubKey())
	addr1 = sdk.ValAddress(pk1.Address())
	addr2 = sdk.ValAddress(pk2.Address())
	addr3 = sdk.ValAddress(pk3.Address())

	emptyAddr   sdk.ValAddress
	emptyPubkey = sdk.NewEmptyConsPubKey()
)
