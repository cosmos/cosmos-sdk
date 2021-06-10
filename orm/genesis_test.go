package orm_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/orm"
	"github.com/cosmos/cosmos-sdk/orm/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestImportExportTableData(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := sdk.NewKVStoreKey("test")
	const prefix = iota
	table := orm.NewAutoUInt64TableBuilder(prefix, 0x1, storeKey, &testdata.GroupInfo{}, cdc).Build()

	ctx := orm.NewMockContext()

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

	err := orm.ImportTableData(ctx, table, groups, 2)
	require.NoError(t, err)

	for _, g := range groups {
		var loaded testdata.GroupInfo
		_, err := table.GetOne(ctx, g.GroupId, &loaded)
		require.NoError(t, err)

		require.Equal(t, g, &loaded)
	}

	var exported []*testdata.GroupInfo
	seq, err := orm.ExportTableData(ctx, table, &exported)
	require.NoError(t, err)
	require.Equal(t, seq, uint64(2))

	for i, g := range exported {
		require.Equal(t, g, groups[i])
	}
}
