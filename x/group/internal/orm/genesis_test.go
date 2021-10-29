package orm

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestImportExportTableData(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := sdk.NewKVStoreKey("test")
	const prefix = iota
	table, err := NewAutoUInt64Table(AutoUInt64TableTablePrefix, 0x1, storeKey, &testdata.GroupInfo{}, cdc)
	require.NoError(t, err)

	ctx := NewMockContext()

	groups := []*testdata.GroupInfo{
		{
			GroupId: 1,
			Admin:   sdk.AccAddress([]byte("admin1-address")),
		},
		{
			GroupId: 2,
			Admin:   sdk.AccAddress([]byte("admin2-address")),
		},
	}

	err = table.Import(ctx, groups, 2)
	require.NoError(t, err)

	for _, g := range groups {
		var loaded testdata.GroupInfo
		_, err := table.GetOne(ctx, g.GroupId, &loaded)
		require.NoError(t, err)

		require.Equal(t, g, &loaded)
	}

	var exported []*testdata.GroupInfo
	seq, err := table.Export(ctx, &exported)
	require.NoError(t, err)
	require.Equal(t, seq, uint64(2))

	for i, g := range exported {
		require.Equal(t, g, groups[i])
	}
}
