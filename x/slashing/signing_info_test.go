package slashing

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetSetValidatorSigningInfo(t *testing.T) {
	ctx, _, _, keeper := createTestInput(t)
	info, found := keeper.getValidatorSigningInfo(ctx, sdk.ValAddress(addrs[0]))
	require.False(t, found)
	newInfo := ValidatorSigningInfo{
		StartHeight:         int64(4),
		IndexOffset:         int64(3),
		JailedUntil:         int64(2),
		SignedBlocksCounter: int64(10),
	}
	keeper.setValidatorSigningInfo(ctx, sdk.ValAddress(addrs[0]), newInfo)
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ValAddress(addrs[0]))
	require.True(t, found)
	require.Equal(t, info.StartHeight, int64(4))
	require.Equal(t, info.IndexOffset, int64(3))
	require.Equal(t, info.JailedUntil, int64(2))
	require.Equal(t, info.SignedBlocksCounter, int64(10))
}

func TestGetSetValidatorSigningBitArray(t *testing.T) {
	ctx, _, _, keeper := createTestInput(t)
	signed := keeper.getValidatorSigningBitArray(ctx, sdk.ValAddress(addrs[0]), 0)
	require.False(t, signed) // treat empty key as unsigned
	keeper.setValidatorSigningBitArray(ctx, sdk.ValAddress(addrs[0]), 0, true)
	signed = keeper.getValidatorSigningBitArray(ctx, sdk.ValAddress(addrs[0]), 0)
	require.True(t, signed) // now should be signed
}
