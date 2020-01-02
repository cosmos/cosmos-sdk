package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/types"
)

func TestGetSetValidatorSigningInfo(t *testing.T) {
	ctx, _, _, _, keeper := CreateTestInput(t, types.DefaultParams())
	info, found := keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(Addrs[0]))
	require.False(t, found)
	newInfo := types.NewValidatorSigningInfo(
		sdk.ConsAddress(Addrs[0]),
		int64(4),
		int64(3),
		time.Unix(2, 0),
		false,
		int64(10),
	)
	keeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(Addrs[0]), newInfo)
	info, found = keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(Addrs[0]))
	require.True(t, found)
	require.Equal(t, info.StartHeight, int64(4))
	require.Equal(t, info.IndexOffset, int64(3))
	require.Equal(t, info.JailedUntil, time.Unix(2, 0).UTC())
	require.Equal(t, info.MissedBlocksCounter, int64(10))
}

func TestValidatorMissedBlockBitArrayEmpty(t *testing.T) {
	ctx, _, _, _, keeper := CreateTestInput(t, types.DefaultParams())
	require.Nil(t, keeper.GetVoteArray(ctx, sdk.ConsAddress(Addrs[0])))
}

func TestValidatorMissedBlockBitArrrayPersisted(t *testing.T) {
	ctx, _, _, _, keeper := CreateTestInput(t, types.DefaultParams())
	size := int64(100)
	array := types.NewVoteArray(size)
	addr := sdk.ConsAddress(Addrs[0])

	misses := map[int64]struct{}{1: struct{}{}, 10: struct{}{}, 29: struct{}{}}
	for v := range misses {
		array.Get(v).Miss()
	}
	keeper.SetVoteArray(ctx, addr, array)

	array = keeper.GetVoteArray(ctx, addr)
	for i := int64(0); i < size; i++ {
		if _, exist := misses[i]; exist {
			require.True(t, array.Get(i).Missed())
		} else {
			require.True(t, array.Get(i).Voted())
		}
	}
}

func TestTombstoned(t *testing.T) {
	ctx, _, _, _, keeper := CreateTestInput(t, types.DefaultParams())
	require.Panics(t, func() { keeper.Tombstone(ctx, sdk.ConsAddress(Addrs[0])) })
	require.False(t, keeper.IsTombstoned(ctx, sdk.ConsAddress(Addrs[0])))

	newInfo := types.NewValidatorSigningInfo(
		sdk.ConsAddress(Addrs[0]),
		int64(4),
		int64(3),
		time.Unix(2, 0),
		false,
		int64(10),
	)
	keeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(Addrs[0]), newInfo)

	require.False(t, keeper.IsTombstoned(ctx, sdk.ConsAddress(Addrs[0])))
	keeper.Tombstone(ctx, sdk.ConsAddress(Addrs[0]))
	require.True(t, keeper.IsTombstoned(ctx, sdk.ConsAddress(Addrs[0])))
	require.Panics(t, func() { keeper.Tombstone(ctx, sdk.ConsAddress(Addrs[0])) })
}

func TestJailUntil(t *testing.T) {
	ctx, _, _, _, keeper := CreateTestInput(t, types.DefaultParams())
	require.Panics(t, func() { keeper.JailUntil(ctx, sdk.ConsAddress(Addrs[0]), time.Now()) })

	newInfo := types.NewValidatorSigningInfo(
		sdk.ConsAddress(Addrs[0]),
		int64(4),
		int64(3),
		time.Unix(2, 0),
		false,
		int64(10),
	)
	keeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(Addrs[0]), newInfo)
	keeper.JailUntil(ctx, sdk.ConsAddress(Addrs[0]), time.Unix(253402300799, 0).UTC())

	info, ok := keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(Addrs[0]))
	require.True(t, ok)
	require.Equal(t, time.Unix(253402300799, 0).UTC(), info.JailedUntil)
}
