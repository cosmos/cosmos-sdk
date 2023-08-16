package types_test

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	keysPK1   = ed25519.GenPrivKeyFromSecret([]byte{1}).PubKey()
	keysPK2   = ed25519.GenPrivKeyFromSecret([]byte{2}).PubKey()
	keysPK3   = ed25519.GenPrivKeyFromSecret([]byte{3}).PubKey()
	keysAddr1 = keysPK1.Address()
	keysAddr2 = keysPK2.Address()
	keysAddr3 = keysPK3.Address()
)

func TestGetValidatorPowerRank(t *testing.T) {
	valAddr1 := sdk.ValAddress(keysAddr1)
	val1 := newValidator(t, valAddr1, keysPK1)
	val1.Tokens = math.ZeroInt()
	val2, val3, val4 := val1, val1, val1
	val2.Tokens = sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)
	val3.Tokens = sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	x := new(big.Int).Exp(big.NewInt(2), big.NewInt(40), big.NewInt(0))
	val4.Tokens = sdk.TokensFromConsensusPower(x.Int64(), sdk.DefaultPowerReduction)

	tests := []struct {
		validator types.Validator
		wantHex   string
	}{
		{val1, "230000000000000000148032c4969889872a13fdbdc46490ea904bc63eba"},
		{val2, "230000000000000001148032c4969889872a13fdbdc46490ea904bc63eba"},
		{val3, "23000000000000000a148032c4969889872a13fdbdc46490ea904bc63eba"},
		{val4, "230000010000000000148032c4969889872a13fdbdc46490ea904bc63eba"},
	}
	for i, tt := range tests {
		got := hex.EncodeToString(types.GetValidatorsByPowerIndexKey(tt.validator, sdk.DefaultPowerReduction, address.NewBech32Codec("cosmosvaloper")))

		require.Equal(t, tt.wantHex, got, "Keys did not match on test case %d", i)
	}
}

func TestGetREDByValDstIndexKey(t *testing.T) {
	tests := []struct {
		delAddr    sdk.AccAddress
		valSrcAddr sdk.ValAddress
		valDstAddr sdk.ValAddress
		wantHex    string
	}{
		{
			sdk.AccAddress(keysAddr1), sdk.ValAddress(keysAddr1), sdk.ValAddress(keysAddr1),
			"36147fcd3b69677678d5ec02423b9b6f156fb439c145147fcd3b69677678d5ec02423b9b6f156fb439c145147fcd3b69677678d5ec02423b9b6f156fb439c145",
		},
		{
			sdk.AccAddress(keysAddr1), sdk.ValAddress(keysAddr2), sdk.ValAddress(keysAddr3),
			"3614e5196ed685cc732ce4baa6aa488af060db48993d147fcd3b69677678d5ec02423b9b6f156fb439c145144c69a7a884af666461d2453b57c43bc3d60b30e1",
		},
		{
			sdk.AccAddress(keysAddr2), sdk.ValAddress(keysAddr1), sdk.ValAddress(keysAddr3),
			"3614e5196ed685cc732ce4baa6aa488af060db48993d144c69a7a884af666461d2453b57c43bc3d60b30e1147fcd3b69677678d5ec02423b9b6f156fb439c145",
		},
	}
	for i, tt := range tests {
		got := hex.EncodeToString(types.GetREDByValDstIndexKey(tt.delAddr, tt.valSrcAddr, tt.valDstAddr))

		require.Equal(t, tt.wantHex, got, "Keys did not match on test case %d", i)
	}
}

func TestGetREDByValSrcIndexKey(t *testing.T) {
	tests := []struct {
		delAddr    sdk.AccAddress
		valSrcAddr sdk.ValAddress
		valDstAddr sdk.ValAddress
		wantHex    string
	}{
		{
			sdk.AccAddress(keysAddr1), sdk.ValAddress(keysAddr1), sdk.ValAddress(keysAddr1),
			"35147fcd3b69677678d5ec02423b9b6f156fb439c145147fcd3b69677678d5ec02423b9b6f156fb439c145147fcd3b69677678d5ec02423b9b6f156fb439c145",
		},
		{
			sdk.AccAddress(keysAddr1), sdk.ValAddress(keysAddr2), sdk.ValAddress(keysAddr3),
			"35144c69a7a884af666461d2453b57c43bc3d60b30e1147fcd3b69677678d5ec02423b9b6f156fb439c14514e5196ed685cc732ce4baa6aa488af060db48993d",
		},
		{
			sdk.AccAddress(keysAddr2), sdk.ValAddress(keysAddr1), sdk.ValAddress(keysAddr3),
			"35147fcd3b69677678d5ec02423b9b6f156fb439c145144c69a7a884af666461d2453b57c43bc3d60b30e114e5196ed685cc732ce4baa6aa488af060db48993d",
		},
	}
	for i, tt := range tests {
		got := hex.EncodeToString(types.GetREDByValSrcIndexKey(tt.delAddr, tt.valSrcAddr, tt.valDstAddr))

		require.Equal(t, tt.wantHex, got, "Keys did not match on test case %d", i)
	}
}

func TestGetValidatorQueueKey(t *testing.T) {
	ts := time.Now()
	height := int64(1024)

	bz := types.GetValidatorQueueKey(ts, height)
	rTs, rHeight, err := types.ParseValidatorQueueKey(bz)
	require.NoError(t, err)
	require.Equal(t, ts.UTC(), rTs.UTC())
	require.Equal(t, rHeight, height)
}

func TestTestGetValidatorQueueKeyOrder(t *testing.T) {
	ts := time.Now().UTC()
	height := int64(1000)

	endKey := types.GetValidatorQueueKey(ts, height)

	keyA := types.GetValidatorQueueKey(ts.Add(-10*time.Minute), height-10)
	keyB := types.GetValidatorQueueKey(ts.Add(-5*time.Minute), height+50)
	keyC := types.GetValidatorQueueKey(ts.Add(10*time.Minute), height+100)

	require.Equal(t, -1, bytes.Compare(keyA, endKey)) // keyA <= endKey
	require.Equal(t, -1, bytes.Compare(keyB, endKey)) // keyB <= endKey
	require.Equal(t, 1, bytes.Compare(keyC, endKey))  // keyB >= endKey
}
