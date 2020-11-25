package types_test

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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
	val1.Tokens = sdk.ZeroInt()
	val2, val3, val4 := val1, val1, val1
	val2.Tokens = sdk.TokensFromConsensusPower(1)
	val3.Tokens = sdk.TokensFromConsensusPower(10)
	x := new(big.Int).Exp(big.NewInt(2), big.NewInt(40), big.NewInt(0))
	val4.Tokens = sdk.TokensFromConsensusPower(x.Int64())

	tests := []struct {
		validator types.Validator
		wantHex   string
	}{
		{val1, "2300000000000000009c288ede7df62742fc3b7d0962045a8cef0f79f6"},
		{val2, "2300000000000000019c288ede7df62742fc3b7d0962045a8cef0f79f6"},
		{val3, "23000000000000000a9c288ede7df62742fc3b7d0962045a8cef0f79f6"},
		{val4, "2300000100000000009c288ede7df62742fc3b7d0962045a8cef0f79f6"},
	}
	for i, tt := range tests {
		got := hex.EncodeToString(types.GetValidatorsByPowerIndexKey(tt.validator))

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
		{sdk.AccAddress(keysAddr1), sdk.ValAddress(keysAddr1), sdk.ValAddress(keysAddr1),
			"3663d771218209d8bd03c482f69dfba57310f0860963d771218209d8bd03c482f69dfba57310f0860963d771218209d8bd03c482f69dfba57310f08609"},
		{sdk.AccAddress(keysAddr1), sdk.ValAddress(keysAddr2), sdk.ValAddress(keysAddr3),
			"363ab62f0d93849be495e21e3e9013a517038f45bd63d771218209d8bd03c482f69dfba57310f086095ef3b5f25c54946d4a89fc0d09d2f126614540f2"},
		{sdk.AccAddress(keysAddr2), sdk.ValAddress(keysAddr1), sdk.ValAddress(keysAddr3),
			"363ab62f0d93849be495e21e3e9013a517038f45bd5ef3b5f25c54946d4a89fc0d09d2f126614540f263d771218209d8bd03c482f69dfba57310f08609"},
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
		{sdk.AccAddress(keysAddr1), sdk.ValAddress(keysAddr1), sdk.ValAddress(keysAddr1),
			"3563d771218209d8bd03c482f69dfba57310f0860963d771218209d8bd03c482f69dfba57310f0860963d771218209d8bd03c482f69dfba57310f08609"},
		{sdk.AccAddress(keysAddr1), sdk.ValAddress(keysAddr2), sdk.ValAddress(keysAddr3),
			"355ef3b5f25c54946d4a89fc0d09d2f126614540f263d771218209d8bd03c482f69dfba57310f086093ab62f0d93849be495e21e3e9013a517038f45bd"},
		{sdk.AccAddress(keysAddr2), sdk.ValAddress(keysAddr1), sdk.ValAddress(keysAddr3),
			"3563d771218209d8bd03c482f69dfba57310f086095ef3b5f25c54946d4a89fc0d09d2f126614540f23ab62f0d93849be495e21e3e9013a517038f45bd"},
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
