package types_test

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	testdatav2 "github.com/cosmos/cosmos-sdk/testutil/testdata/v2"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMarshalUnmarshalGogo(t *testing.T) {
	registry := testdata.NewTestInterfaceRegistry()

	doggo := &testdata.Dog{
		Size_: "Really Big",
		Name:  "Clifford",
	}
	any, err := types.NewAnyWithValue(doggo)
	require.NoError(t, err)
	require.NotNil(t, any)

	jm := &jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: false,
		Indent:       "   ",
		OrigName:     true,
		AnyResolver:  registry,
	}

	bz, err := any.MarshalJSONPB(jm)
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	any2, err := types.NewAnyWithValue(&testdata.Dog{})
	require.NoError(t, err)
	jum := &jsonpb.Unmarshaler{
		AllowUnknownFields: true,
		AnyResolver:        registry,
	}

	require.NoError(t, any2.UnmarshalJSONPB(jum, bz))
}

func TestMarshalUnmarshalV2(t *testing.T) {
	registry := testdata.NewTestInterfaceRegistry()

	snakey := &testdatav2.Snake{
		Name: "Elong",
		Age:  32,
	}
	any, err := types.NewAnyWithValue(snakey)
	require.NoError(t, err)
	require.NotNil(t, any)

	jm := &jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: false,
		Indent:       "   ",
		OrigName:     true,
		AnyResolver:  registry,
	}

	bz, err := any.MarshalJSONPB(jm)
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	snakey2 := &testdatav2.Snake{}
	any, err = types.NewAnyWithValue(snakey2)
	require.NoError(t, err)
	require.NotNil(t, any)

	jum := &jsonpb.Unmarshaler{
		AllowUnknownFields: false,
		AnyResolver:        registry,
	}
	require.NoError(t, any.UnmarshalJSONPB(jum, bz))
}
