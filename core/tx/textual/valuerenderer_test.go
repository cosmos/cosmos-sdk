package textual_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/core/tx/textual"
	"cosmossdk.io/core/tx/textual/internal/testpb"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestFormatInteger(t *testing.T) {
	type integerTest []string
	var testcases []integerTest
	raw, err := ioutil.ReadFile("./internal/fixtures/integers.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	r := textual.NewADR050ValueRenderer()

	for _, tc := range testcases {
		// Test integers as protobuf uint64
		b, err := strconv.ParseUint(tc[0], 10, 64)
		if err == nil {
			output, err := formatGoType(r, b)
			require.NoError(t, err)

			require.Equal(t, []string{tc[1]}, output)
		}

		// Test integers as protobuf uint32
		b, err = strconv.ParseUint(tc[0], 10, 32)
		if err == nil {
			output, err := formatGoType(r, b)
			require.NoError(t, err)

			require.Equal(t, []string{tc[1]}, output)
		}

		// Test integers as sdk.Ints
		i, ok := math.NewIntFromString(tc[0])
		if ok {
			output, err := formatGoType(r, i)
			require.NoError(t, err)

			require.Equal(t, []string{tc[1]}, output)
		}
	}
}

func TestFormatDecimal(t *testing.T) {
	type decimalTest []string
	var testcases []decimalTest
	raw, err := ioutil.ReadFile("./internal/fixtures/decimals.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	r := textual.NewADR050ValueRenderer()

	for _, tc := range testcases {
		d, err := sdk.NewDecFromStr(tc[0])
		require.NoError(t, err)
		output, err := formatGoType(r, d)
		require.NoError(t, err)

		require.Equal(t, []string{tc[1]}, output)
	}
}

// func TestFormatCoin(t *testing.T) {
// 	var testcases []coinTest
// 	raw, err := ioutil.ReadFile("./internal/fixtures/coins.json")
// 	require.NoError(t, err)
// 	err = json.Unmarshal(raw, &testcases)
// 	require.NoError(t, err)

// 	r :=textual.NewADR050ValueRenderer()

// 	for _, tc := range testcases {
// 		output, err := formatCoin(tc.coin, bank.Metadata{
// 			Display:    tc.metadata.Denom,
// 			DenomUnits: []*bank.DenomUnit{{Denom: tc.coin.Denom, Exponent: 0}, {Denom: tc.metadata.Denom, Exponent: tc.metadata.Exponent}},
// 		})
// 		require.NoError(t, err)

// 		require.Equal(t, tc.expRes, output)
// 	}
// }

// func TestFormatCoins(t *testing.T) {
// 	var testcases []coinTest
// 	raw, err := ioutil.ReadFile("./internal/fixtures/coins.json")
// 	require.NoError(t, err)
// 	err = json.Unmarshal(raw, &testcases)
// 	require.NoError(t, err)

// 	for _, tc := range testcases {
// 		output, err := formatCoins(sdk.NewCoins(tc.coin), bank.Metadata{
// 			Display:    tc.metadata.Denom,
// 			DenomUnits: []*bank.DenomUnit{{Denom: tc.coin.Denom, Exponent: 0}, {Denom: tc.metadata.Denom, Exponent: tc.metadata.Exponent}},
// 		})
// 		require.NoError(t, err)

// 		require.Equal(t, tc.expRes, output)
// 	}
// }

type coinTestMetadata struct {
	Denom    string `json:"denom"`
	Exponent uint32 `json:"exponent"`
}

type coinTest struct {
	coin     sdk.Coin
	metadata coinTestMetadata
	expRes   string
}

func (t *coinTest) UnmarshalJSON(b []byte) error {
	a := []interface{}{&t.coin, &t.metadata, &t.expRes}
	return json.Unmarshal(b, &a)
}

// formatGoType is like ValueRenderer's Format(), but taking a Go type as input
// value.
func formatGoType(r textual.ValueRenderer, v interface{}) ([]string, error) {
	a := testpb.A{}

	switch v := v.(type) {
	case uint32:
		return r.Format(context.Background(), a.ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name("UINT32")), protoreflect.ValueOf(v))
	case uint64:
		return r.Format(context.Background(), a.ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name("UINT64")), protoreflect.ValueOf(v))
	case int32:
		return r.Format(context.Background(), a.ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name("INT32")), protoreflect.ValueOf(v))
	case int64:
		return r.Format(context.Background(), a.ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name("INT64")), protoreflect.ValueOf(v))
	case math.Int:
		return r.Format(context.Background(), a.ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name("SDKINT")), protoreflect.ValueOf(v.String()))
	case sdk.Dec:
		return r.Format(context.Background(), a.ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name("SDKDEC")), protoreflect.ValueOf(v.String()))

	default:
		return nil, fmt.Errorf("value %s of type %T not recognized", v, v)
	}
}
