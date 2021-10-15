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

	obj := testdata.TableModel{
		Id:   1,
		Name: "Some name",
	}
	err = tb.Create(store, &obj)
	require.NoError(t, err)

	specs := map[string]struct {
		src PrimaryKeyed
		exp bool
	}{

		"same object": {src: &obj, exp: true},
		"clone": {
			src: &testdata.TableModel{
				Id:   1,
				Name: "Some name",
			},
			exp: true,
		},
		"different primary key": {
			src: &testdata.TableModel{
				Id:   2,
				Name: "Some name",
			},
			exp: false,
		},
		"different type, same key": {
			src: mockPrimaryKeyed{&obj},
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
	*testdata.TableModel
}
