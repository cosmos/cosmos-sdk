package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store/types"
)

func TestPruningOptions_FlushVersion(t *testing.T) {
	t.Parallel()
	require.True(t, types.PruneEverything.FlushVersion(-1))
	require.True(t, types.PruneEverything.FlushVersion(0))
	require.True(t, types.PruneEverything.FlushVersion(1))
	require.True(t, types.PruneEverything.FlushVersion(2))

	require.True(t, types.PruneNothing.FlushVersion(-1))
	require.True(t, types.PruneNothing.FlushVersion(0))
	require.True(t, types.PruneNothing.FlushVersion(1))
	require.True(t, types.PruneNothing.FlushVersion(2))

	require.False(t, types.PruneSyncable.FlushVersion(-1))
	require.True(t, types.PruneSyncable.FlushVersion(0))
	require.False(t, types.PruneSyncable.FlushVersion(1))
	require.True(t, types.PruneSyncable.FlushVersion(100))
	require.False(t, types.PruneSyncable.FlushVersion(101))
}

func TestPruningOptions_SnapshotVersion(t *testing.T) {
	t.Parallel()
	require.False(t, types.PruneEverything.SnapshotVersion(-1))
	require.False(t, types.PruneEverything.SnapshotVersion(0))
	require.False(t, types.PruneEverything.SnapshotVersion(1))
	require.False(t, types.PruneEverything.SnapshotVersion(2))

	require.True(t, types.PruneNothing.SnapshotVersion(-1))
	require.True(t, types.PruneNothing.SnapshotVersion(0))
	require.True(t, types.PruneNothing.SnapshotVersion(1))
	require.True(t, types.PruneNothing.SnapshotVersion(2))

	require.False(t, types.PruneSyncable.SnapshotVersion(-1))
	require.True(t, types.PruneSyncable.SnapshotVersion(0))
	require.False(t, types.PruneSyncable.SnapshotVersion(1))
	require.True(t, types.PruneSyncable.SnapshotVersion(10000))
	require.False(t, types.PruneSyncable.SnapshotVersion(10001))
}

func TestPruningOptions_IsValid(t *testing.T) {
	t.Parallel()
	type fields struct {
		KeepEvery     int64
		SnapshotEvery int64
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"PruneEverything", fields{types.PruneEverything.KeepEvery, types.PruneEverything.SnapshotEvery}, true},
		{"PruneNothing", fields{types.PruneNothing.KeepEvery, types.PruneNothing.SnapshotEvery}, true},
		{"PruneSyncable", fields{types.PruneSyncable.KeepEvery, types.PruneSyncable.SnapshotEvery}, true},
		{"KeepEvery=0", fields{0, 0}, false},
		{"KeepEvery<0", fields{-1, 0}, false},
		{"SnapshotEvery<0", fields{1, -1}, false},
		{"SnapshotEvery%KeepEvery!=0", fields{15, 30}, true},
		{"SnapshotEvery%KeepEvery!=0", fields{15, 20}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			po := types.PruningOptions{
				KeepEvery:     tt.fields.KeepEvery,
				SnapshotEvery: tt.fields.SnapshotEvery,
			}
			require.Equal(t, tt.want, po.IsValid(), "IsValid() = %v, want %v", po.IsValid(), tt.want)
		})
	}
}
