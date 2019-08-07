package simapp

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	delPk1    = ed25519.GenPrivKey().PubKey()
	delAddr1  = sdk.AccAddress(delPk1.Address())
	valAddr1  = sdk.ValAddress(delPk1.Address())
	consAddr1 = sdk.ConsAddress(delPk1.Address().Bytes())
)

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	auth.RegisterCodec(cdc)
	distr.RegisterCodec(cdc)
	gov.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	supply.RegisterCodec(cdc)
	return
}

func TestGetSimulationLog(t *testing.T) {
	cdc := makeTestCodec()

	tests := []struct {
		store   string
		kvPairs []cmn.KVPair
	}{
		{auth.StoreKey, []cmn.KVPair{
			{Key: auth.AddressStoreKey(delAddr1), Value: cdc.MustMarshalBinaryBare(auth.BaseAccount{})},
			{Key: auth.AddressStoreKey(delAddr1), Value: cdc.MustMarshalBinaryBare(auth.BaseAccount{})},
		}},
		{mint.StoreKey, []cmn.KVPair{
			{Key: mint.MinterKey, Value: cdc.MustMarshalBinaryLengthPrefixed(mint.Minter{})},
			{Key: mint.MinterKey, Value: cdc.MustMarshalBinaryLengthPrefixed(mint.Minter{})},
		}},
		{staking.StoreKey, []cmn.KVPair{
			{Key: staking.LastValidatorPowerKey, Value: valAddr1.Bytes()},
			{Key: staking.LastValidatorPowerKey, Value: valAddr1.Bytes()},
		}},
		{gov.StoreKey, []cmn.KVPair{
			{Key: gov.VoteKey(1, delAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(gov.Vote{})},
			{Key: gov.VoteKey(1, delAddr1), Value: cdc.MustMarshalBinaryLengthPrefixed(gov.Vote{})},
		}},
		{distribution.StoreKey, []cmn.KVPair{
			{Key: distr.ProposerKey, Value: consAddr1.Bytes()},
			{Key: distr.ProposerKey, Value: consAddr1.Bytes()},
		}},
		{slashing.StoreKey, []cmn.KVPair{
			{Key: slashing.GetValidatorMissedBlockBitArrayKey(consAddr1, 6), Value: cdc.MustMarshalBinaryLengthPrefixed(true)},
			{Key: slashing.GetValidatorMissedBlockBitArrayKey(consAddr1, 6), Value: cdc.MustMarshalBinaryLengthPrefixed(true)},
		}},
		{supply.StoreKey, []cmn.KVPair{
			{Key: supply.SupplyKey, Value: cdc.MustMarshalBinaryLengthPrefixed(supply.NewSupply(sdk.Coins{}))},
			{Key: supply.SupplyKey, Value: cdc.MustMarshalBinaryLengthPrefixed(supply.NewSupply(sdk.Coins{}))},
		}},
		{"Empty", []cmn.KVPair{{}, {}}},
		{"OtherStore", []cmn.KVPair{
			{Key: []byte("key"), Value: []byte("value")},
			{Key: []byte("key"), Value: []byte("other_value")},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.store, func(t *testing.T) {
			require.NotPanics(t, func() { GetSimulationLog(tt.store, make(sdk.StoreDecoderRegistry) , cdc, tt.kvPairs) }, tt.store)
		})
	}
}
