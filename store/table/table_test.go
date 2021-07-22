package table

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTableBuilder(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	const anyPrefix = 0x10

	specs := map[string]struct {
		model       codec.ProtoMarshaler
		idxKeyCodec IndexKeyCodec
		expPanic    bool
	}{
		"happy path": {
			model:       &testdata.TableModel{},
			idxKeyCodec: Max255DynamicLengthIndexKeyCodec{},
		},
		"nil model": {
			idxKeyCodec: Max255DynamicLengthIndexKeyCodec{},
			expPanic:    true,
		},
		"nil idxKeyCodec": {
			model:    &testdata.TableModel{},
			expPanic: true,
		},
	}

	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			f := func() {
				NewTableBuilder(anyPrefix, spec.model, spec.idxKeyCodec, cdc)
			}
			if spec.expPanic {
				require.Panics(t, f)
			} else {
				require.NotPanics(t, f)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	specs := map[string]struct {
		src    codec.ProtoMarshaler
		expErr *errors.Error
	}{
		"happy path": {
			src: &testdata.TableModel{
				Id:   "my-id",
				Name: "some name",
			},
		},
		"wrong type": {
			src: &testdata.Cat{
				Moniker: "cat moniker",
				Lives:   10,
			},
			expErr: ErrType,
		},
		"model validation fails": {
			src: &testdata.TableModel{
				Id:   "my-id",
				Name: "",
			},
			expErr: testdata.ErrTest,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			interfaceRegistry := types.NewInterfaceRegistry()
			cdc := codec.NewProtoCodec(interfaceRegistry)

			ctx := NewMockContext()
			store := ctx.KVStore(sdk.NewKVStoreKey("test"))

			const anyPrefix = 0x10
			tableBuilder := NewTableBuilder(anyPrefix, &testdata.TableModel{}, Max255DynamicLengthIndexKeyCodec{}, cdc)
			myTable := tableBuilder.Build()

			err := myTable.Create(store, []byte("my-id"), spec.src)

			require.True(t, spec.expErr.Is(err), err)
			shouldExists := spec.expErr == nil
			assert.Equal(t, shouldExists, myTable.Has(store, []byte("my-id")), fmt.Sprintf("expected %v", shouldExists))

			// then
			var loaded testdata.TableModel
			err = myTable.GetOne(store, []byte("my-id"), &loaded)
			if spec.expErr != nil {
				require.True(t, ErrNotFound.Is(err))
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
			src: &testdata.TableModel{
				Id:   "my-id",
				Name: "some name",
			},
		},
		"wrong type": {
			src: &testdata.Cat{
				Moniker: "cat moniker",
				Lives:   10,
			},
			expErr: ErrType,
		},
		"model validation fails": {
			src: &testdata.TableModel{
				Id:   "my-id",
				Name: "",
			},
			expErr: testdata.ErrTest,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			interfaceRegistry := types.NewInterfaceRegistry()
			cdc := codec.NewProtoCodec(interfaceRegistry)

			ctx := NewMockContext()
			store := ctx.KVStore(sdk.NewKVStoreKey("test"))

			const anyPrefix = 0x10
			tableBuilder := NewTableBuilder(anyPrefix, &testdata.TableModel{}, Max255DynamicLengthIndexKeyCodec{}, cdc)
			myTable := tableBuilder.Build()

			initValue := testdata.TableModel{
				Id:   "my-id",
				Name: "old name",
			}

			err := myTable.Create(store, []byte("my-id"), &initValue)
			require.NoError(t, err)

			// when
			err = myTable.Save(store, []byte("my-id"), spec.src)
			require.True(t, spec.expErr.Is(err), "got ", err)

			// then
			var loaded testdata.TableModel
			require.NoError(t, myTable.GetOne(store, []byte("my-id"), &loaded))
			if spec.expErr == nil {
				assert.Equal(t, spec.src, &loaded)
			} else {
				assert.Equal(t, initValue, loaded)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	specs := map[string]struct {
		rowId  []byte
		expErr *errors.Error
	}{
		"happy path": {
			rowId: []byte("my-id"),
		},
		"not found": {
			rowId:  []byte("not-found"),
			expErr: ErrNotFound,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			interfaceRegistry := types.NewInterfaceRegistry()
			cdc := codec.NewProtoCodec(interfaceRegistry)

			ctx := NewMockContext()
			store := ctx.KVStore(sdk.NewKVStoreKey("test"))

			const anyPrefix = 0x10
			tableBuilder := NewTableBuilder(anyPrefix, &testdata.TableModel{}, Max255DynamicLengthIndexKeyCodec{}, cdc)
			myTable := tableBuilder.Build()

			initValue := testdata.TableModel{
				Id:   "my-id",
				Name: "some name",
			}

			err := myTable.Create(store, []byte("my-id"), &initValue)
			require.NoError(t, err)

			// when
			err = myTable.Delete(store, spec.rowId)
			require.True(t, spec.expErr.Is(err), "got ", err)

			// then
			var loaded testdata.TableModel
			if spec.expErr == ErrNotFound {
				require.NoError(t, myTable.GetOne(store, []byte("my-id"), &loaded))
				assert.Equal(t, initValue, loaded)
			} else {
				err := myTable.GetOne(store, []byte("my-id"), &loaded)
				require.Error(t, err)
				require.Equal(t, err, ErrNotFound)
			}
		})
	}
}
