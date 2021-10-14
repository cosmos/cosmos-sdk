package orm

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func TestContains(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	ctx := NewMockContext()
	store := ctx.KVStore(sdk.NewKVStoreKey("test"))

	builder, err := NewPrimaryKeyTableBuilder([2]byte{0x1}, &testdata.TableModel{}, cdc)
	require.NoError(t, err)
	tb := builder.Build()

	myPersistentObj := testdata.TableModel{
		Id:   1,
		Name: "Some name",
	}
	err = tb.Create(store, &myPersistentObj)
	require.NoError(t, err)

	specs := map[string]struct {
		src PrimaryKeyed
		exp bool
	}{

		"same object": {src: &myPersistentObj, exp: true},
		"clone": {
			src: &testdata.GroupMember{
				Group:  []byte("group-a"),
				Member: []byte("member-one"),
				Weight: 1,
			},
			exp: true,
		},
		"different primary key": {
			src: &testdata.GroupMember{
				Group:  []byte("another group"),
				Member: []byte("member-one"),
				Weight: 1,
			},
			exp: false,
		},
		"different type, same key": {
			src: mockPrimaryKeyed{&myPersistentObj},
			exp: false,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			got := tb.Contains(store, spec.src)
			assert.Equal(t, spec.exp, got)
		})
	}
}

type mockPrimaryKeyed struct {
	*testdata.GroupMember
}
