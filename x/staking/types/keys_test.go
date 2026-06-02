package types_test

import (
	"bytes"
	"encoding/hex"
	math2 "math"
	"math/big"
	"strconv"
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
		{val1, "230000000000000000149c288ede7df62742fc3b7d0962045a8cef0f79f6"},
		{val2, "230000000000000001149c288ede7df62742fc3b7d0962045a8cef0f79f6"},
		{val3, "23000000000000000a149c288ede7df62742fc3b7d0962045a8cef0f79f6"},
		{val4, "230000010000000000149c288ede7df62742fc3b7d0962045a8cef0f79f6"},
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
			"361463d771218209d8bd03c482f69dfba57310f086091463d771218209d8bd03c482f69dfba57310f086091463d771218209d8bd03c482f69dfba57310f08609",
		},
		{
			sdk.AccAddress(keysAddr1), sdk.ValAddress(keysAddr2), sdk.ValAddress(keysAddr3),
			"36143ab62f0d93849be495e21e3e9013a517038f45bd1463d771218209d8bd03c482f69dfba57310f08609145ef3b5f25c54946d4a89fc0d09d2f126614540f2",
		},
		{
			sdk.AccAddress(keysAddr2), sdk.ValAddress(keysAddr1), sdk.ValAddress(keysAddr3),
			"36143ab62f0d93849be495e21e3e9013a517038f45bd145ef3b5f25c54946d4a89fc0d09d2f126614540f21463d771218209d8bd03c482f69dfba57310f08609",
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
			"351463d771218209d8bd03c482f69dfba57310f086091463d771218209d8bd03c482f69dfba57310f086091463d771218209d8bd03c482f69dfba57310f08609",
		},
		{
			sdk.AccAddress(keysAddr1), sdk.ValAddress(keysAddr2), sdk.ValAddress(keysAddr3),
			"35145ef3b5f25c54946d4a89fc0d09d2f126614540f21463d771218209d8bd03c482f69dfba57310f08609143ab62f0d93849be495e21e3e9013a517038f45bd",
		},
		{
			sdk.AccAddress(keysAddr2), sdk.ValAddress(keysAddr1), sdk.ValAddress(keysAddr3),
			"351463d771218209d8bd03c482f69dfba57310f08609145ef3b5f25c54946d4a89fc0d09d2f126614540f2143ab62f0d93849be495e21e3e9013a517038f45bd",
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

func TestGetConsKeyRotationQueueKey(t *testing.T) {
	tests := []struct {
		name    string
		ts      time.Time
		valAddr sdk.ValAddress
	}{
		{"keysAddr1 now", time.Now(), sdk.ValAddress(keysAddr1)},
		{"keysAddr2 epoch", time.Unix(0, 0), sdk.ValAddress(keysAddr2)},
		{"keysAddr3 future", time.Now().Add(24 * time.Hour), sdk.ValAddress(keysAddr3)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := types.GetConsKeyRotationQueueKey(tt.ts, tt.valAddr)
			gotTs, gotValAddr, err := types.ParseConsKeyRotationQueueKey(bz)
			require.NoError(t, err)
			require.Equal(t, tt.ts.UTC(), gotTs.UTC())
			require.Equal(t, tt.valAddr, gotValAddr)
		})
	}
}

func TestGetConsKeyRotationQueueKeyOrder(t *testing.T) {
	ts := time.Now().UTC()
	valAddr := sdk.ValAddress(keysAddr1)
	endKey := types.GetConsKeyRotationQueueKey(ts, valAddr)

	keyA := types.GetConsKeyRotationQueueKey(ts.Add(-10*time.Minute), valAddr)
	keyB := types.GetConsKeyRotationQueueKey(ts.Add(-5*time.Minute), valAddr)
	keyC := types.GetConsKeyRotationQueueKey(ts.Add(10*time.Minute), valAddr)

	require.Equal(t, -1, bytes.Compare(keyA, endKey))
	require.Equal(t, -1, bytes.Compare(keyB, endKey))
	require.Equal(t, 1, bytes.Compare(keyC, endKey))
}

func TestParseConsKeyRotationQueueKey(t *testing.T) {
	ts := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	valAddr := sdk.ValAddress(keysAddr1)

	tests := []struct {
		name           string
		buildKey       func() []byte
		expErrContains string
		expTs          time.Time
		expValAddr     sdk.ValAddress
	}{
		{
			name:       "valid key",
			buildKey:   func() []byte { return types.GetConsKeyRotationQueueKey(ts, valAddr) },
			expTs:      ts,
			expValAddr: valAddr,
		},
		{
			name: "wrong prefix",
			buildKey: func() []byte {
				bz := types.GetConsKeyRotationQueueKey(ts, valAddr)
				bz[0] = 0xff
				return bz
			},
			expErrContains: "invalid prefix",
		},
		{
			name: "unparseable time bytes",
			buildKey: func() []byte {
				bz := types.GetConsKeyRotationQueueKey(ts, valAddr)
				prefixLen := len(types.ConsKeyRotationQueueKey)
				for i := prefixLen; i < prefixLen+5; i++ {
					bz[i] = 0xff
				}
				return bz
			},
			expErrContains: "cannot parse",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTs, gotValAddr, err := types.ParseConsKeyRotationQueueKey(tt.buildKey())
			if tt.expErrContains != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expErrContains)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expTs.UTC(), gotTs.UTC())
			require.Equal(t, tt.expValAddr, gotValAddr)
		})
	}
}

func TestGetConsKeyRotationQueueTimePrefix(t *testing.T) {
	ts := time.Now()
	prefix := types.GetConsKeyRotationQueueTimePrefix(ts)
	full := types.GetConsKeyRotationQueueKey(ts, sdk.ValAddress(keysAddr1))

	require.True(t, bytes.HasPrefix(full, prefix))
	require.Equal(t, len(types.ConsKeyRotationQueueKey)+len(sdk.FormatTimeBytes(ts)), len(prefix))
}

func TestGetValidatorConsKeyRotationKey(t *testing.T) {
	tests := []struct {
		valAddr sdk.ValAddress
		wantHex string
	}{
		{sdk.ValAddress(keysAddr1), "921463d771218209d8bd03c482f69dfba57310f08609"},
		{sdk.ValAddress(keysAddr2), "92145ef3b5f25c54946d4a89fc0d09d2f126614540f2"},
		{sdk.ValAddress(keysAddr3), "92143ab62f0d93849be495e21e3e9013a517038f45bd"},
	}
	for i, tt := range tests {
		got := hex.EncodeToString(types.GetValidatorConsKeyRotationKey(tt.valAddr))
		require.Equal(t, tt.wantHex, got, "Keys did not match on test case %d", i)
	}
}

func TestGetRotationLockedConsAddrIndexKey(t *testing.T) {
	tests := []struct {
		consAddr sdk.ConsAddress
		wantHex  string
	}{
		{sdk.ConsAddress(keysAddr1), "931463d771218209d8bd03c482f69dfba57310f08609"},
		{sdk.ConsAddress(keysAddr2), "93145ef3b5f25c54946d4a89fc0d09d2f126614540f2"},
		{sdk.ConsAddress(keysAddr3), "93143ab62f0d93849be495e21e3e9013a517038f45bd"},
	}
	for i, tt := range tests {
		got := hex.EncodeToString(types.GetRotationLockedConsAddrIndexKey(tt.consAddr))
		require.Equal(t, tt.wantHex, got, "Keys did not match on test case %d", i)
	}
}

func TestGetConsKeyRotationApplyQueueKey(t *testing.T) {
	tests := []struct {
		name        string
		applyHeight int64
		valAddr     sdk.ValAddress
	}{
		{"zero height keysAddr1", 0, sdk.ValAddress(keysAddr1)},
		{"small height keysAddr2", 42, sdk.ValAddress(keysAddr2)},
		{"large height keysAddr3", math2.MaxInt64, sdk.ValAddress(keysAddr3)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := types.GetConsKeyRotationApplyQueueKey(tt.applyHeight, tt.valAddr)
			gotHeight, gotValAddr, err := types.ParseConsKeyRotationApplyQueueKey(bz)
			require.NoError(t, err)
			require.Equal(t, tt.applyHeight, gotHeight)
			require.Equal(t, tt.valAddr, gotValAddr)
		})
	}
}

func TestGetConsKeyRotationApplyQueueKeyOrder(t *testing.T) {
	valAddr := sdk.ValAddress(keysAddr1)
	endKey := types.GetConsKeyRotationApplyQueueKey(100, valAddr)

	keyA := types.GetConsKeyRotationApplyQueueKey(50, valAddr)
	keyB := types.GetConsKeyRotationApplyQueueKey(99, valAddr)
	keyC := types.GetConsKeyRotationApplyQueueKey(101, valAddr)

	require.Equal(t, -1, bytes.Compare(keyA, endKey))
	require.Equal(t, -1, bytes.Compare(keyB, endKey))
	require.Equal(t, 1, bytes.Compare(keyC, endKey))
}

func TestGetConsKeyRotationApplyQueueHeightPrefix(t *testing.T) {
	prefix := types.GetConsKeyRotationApplyQueueHeightPrefix(42)
	full := types.GetConsKeyRotationApplyQueueKey(42, sdk.ValAddress(keysAddr1))
	require.True(t, bytes.HasPrefix(full, prefix))
	require.Equal(t, len(types.ConsKeyRotationApplyQueueKey)+8, len(prefix))
}

func TestParseConsKeyRotationApplyQueueKey_Errors(t *testing.T) {
	valAddr := sdk.ValAddress(keysAddr1)
	tests := []struct {
		name           string
		buildKey       func() []byte
		expErrContains string
	}{
		{
			name: "wrong prefix",
			buildKey: func() []byte {
				bz := types.GetConsKeyRotationApplyQueueKey(1, valAddr)
				bz[0] = 0xff
				return bz
			},
			expErrContains: "invalid prefix",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := types.ParseConsKeyRotationApplyQueueKey(tt.buildKey())
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expErrContains)
		})
	}
}

func TestGetHistoricalInfoKey(t *testing.T) {
	tests := []struct {
		height int64
		want   []byte
	}{
		{0, append(types.HistoricalInfoKey, []byte{0, 0, 0, 0, 0, 0, 0, 0}...)},
		{1, append(types.HistoricalInfoKey, []byte{0, 0, 0, 0, 0, 0, 0, 1}...)},
		{2, append(types.HistoricalInfoKey, []byte{0, 0, 0, 0, 0, 0, 0, 2}...)},
		{514, append(types.HistoricalInfoKey, []byte{0, 0, 0, 0, 0, 0, 2, 2}...)},
		{math2.MaxInt64, append(types.HistoricalInfoKey, []byte{127, 255, 255, 255, 255, 255, 255, 255}...)},
	}
	for _, tt := range tests {
		t.Run(strconv.FormatInt(tt.height, 10), func(t *testing.T) {
			require.Equal(t, tt.want, types.GetHistoricalInfoKey(tt.height))
		})
	}
}
