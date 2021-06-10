package orm_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/orm"
	"github.com/cosmos/cosmos-sdk/orm/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

func TestCreate(t *testing.T) {
	specs := map[string]struct {
		src    codec.ProtoMarshaler
		expErr *errors.Error
	}{
		"happy path": {
			src: &testdata.GroupInfo{
				Description: "my group",
				Admin:       sdk.AccAddress([]byte("my-admin-address")),
			},
		},
		"wrong type": {
			src: &testdata.GroupMember{
				Group:  sdk.AccAddress(orm.EncodeSequence(1)),
				Member: sdk.AccAddress([]byte("member-address")),
				Weight: 10,
			},
			expErr: orm.ErrType,
		},
		"model validation fails": {
			src:    &testdata.GroupInfo{Description: "invalid"},
			expErr: testdata.ErrTest,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			interfaceRegistry := types.NewInterfaceRegistry()
			cdc := codec.NewProtoCodec(interfaceRegistry)

			storeKey := sdk.NewKVStoreKey("test")
			const anyPrefix = 0x10
			tableBuilder := orm.NewTableBuilder(anyPrefix, storeKey, &testdata.GroupInfo{}, orm.Max255DynamicLengthIndexKeyCodec{}, cdc)
			myTable := tableBuilder.Build()

			ctx := orm.NewMockContext()
			err := myTable.Create(ctx, []byte("my-id"), spec.src)

			require.True(t, spec.expErr.Is(err), err)
			shouldExists := spec.expErr == nil
			assert.Equal(t, shouldExists, myTable.Has(ctx, []byte("my-id")), fmt.Sprintf("expected %v", shouldExists))

			// then
			var loaded testdata.GroupInfo
			err = myTable.GetOne(ctx, []byte("my-id"), &loaded)
			if spec.expErr != nil {
				require.True(t, orm.ErrNotFound.Is(err))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, spec.src, &loaded)
		})
	}

}
func TestUpdate(t *testing.T) {
	specs := map[string]struct {
		src    codec.ProtoMarshaler
		expErr *errors.Error
	}{
		"happy path": {
			src: &testdata.GroupInfo{
				Description: "my group",
				Admin:       sdk.AccAddress([]byte("my-admin-address")),
			},
		},
		"wrong type": {
			src: &testdata.GroupMember{
				Group:  sdk.AccAddress(orm.EncodeSequence(1)),
				Member: sdk.AccAddress([]byte("member-address")),
				Weight: 9999,
			},
			expErr: orm.ErrType,
		},
		"model validation fails": {
			src:    &testdata.GroupInfo{Description: "invalid"},
			expErr: testdata.ErrTest,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			interfaceRegistry := types.NewInterfaceRegistry()
			cdc := codec.NewProtoCodec(interfaceRegistry)

			storeKey := sdk.NewKVStoreKey("test")
			const anyPrefix = 0x10
			tableBuilder := orm.NewTableBuilder(anyPrefix, storeKey, &testdata.GroupInfo{}, orm.Max255DynamicLengthIndexKeyCodec{}, cdc)
			myTable := tableBuilder.Build()

			initValue := testdata.GroupInfo{
				Description: "my old group description",
				Admin:       sdk.AccAddress([]byte("my-old-admin-address")),
			}
			ctx := orm.NewMockContext()
			err := myTable.Create(ctx, []byte("my-id"), &initValue)
			require.NoError(t, err)

			// when
			err = myTable.Save(ctx, []byte("my-id"), spec.src)
			require.True(t, spec.expErr.Is(err), "got ", err)

			// then
			var loaded testdata.GroupInfo
			require.NoError(t, myTable.GetOne(ctx, []byte("my-id"), &loaded))
			if spec.expErr == nil {
				assert.Equal(t, spec.src, &loaded)
			} else {
				assert.Equal(t, initValue, loaded)
			}
		})
	}

}
