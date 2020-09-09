package types

import (
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// nolint:deadcode,unused
var (
	pk1      = ed25519.GenPrivKey().PubKey()
	pk2      = ed25519.GenPrivKey().PubKey()
	pk3      = ed25519.GenPrivKey().PubKey()
	addr1, _    = sdk.Bech32ifyAddressBytes(sdk.Bech32PrefixAccAddr, pk1.Address().Bytes())
	addr2, _    = sdk.Bech32ifyAddressBytes(sdk.Bech32PrefixAccAddr, pk2.Address().Bytes())
	addr3, _    = sdk.Bech32ifyAddressBytes(sdk.Bech32PrefixAccAddr, pk3.Address().Bytes())
	valAddr1 = sdk.ValAddress(pk1.Address())
	valAddr2 = sdk.ValAddress(pk2.Address())
	valAddr3 = sdk.ValAddress(pk3.Address())

	emptyAddr   sdk.ValAddress
	emptyPubkey crypto.PubKey
)
