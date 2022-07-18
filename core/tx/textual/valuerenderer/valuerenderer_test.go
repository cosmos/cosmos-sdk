package valuerenderer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/core/tx/textual/internal/testpb"
	"cosmossdk.io/core/tx/textual/valuerenderer"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestFormatInteger(t *testing.T) {
	type integerTest []string
	var testcases []integerTest
	raw, err := ioutil.ReadFile("../internal/testdata/integers.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		// Parse test case strings as protobuf uint64
		b, err := strconv.ParseUint(tc[0], 10, 64)
		if err == nil {
			r, err := getVRFromGoType(b)
			require.NoError(t, err)
			output, err := r.Format(context.Background(), protoreflect.ValueOf(b))
			require.NoError(t, err)

			require.Equal(t, []string{tc[1]}, output)
		}

		// Parse test case strings as protobuf uint32
		b, err = strconv.ParseUint(tc[0], 10, 32)
		if err == nil {
			r, err := getVRFromGoType(b)
			require.NoError(t, err)
			output, err := r.Format(context.Background(), protoreflect.ValueOf(b))
			require.NoError(t, err)

			require.Equal(t, []string{tc[1]}, output)
		}

		// Parse test case strings as sdk.Ints
		i, ok := math.NewIntFromString(tc[0])
		if ok {
			r, err := getVRFromGoType(b)
			require.NoError(t, err)
			output, err := r.Format(context.Background(), protoreflect.ValueOf(i))
			require.NoError(t, err)

			require.Equal(t, []string{tc[1]}, output)
		}
	}
}

func TestFormatDecimal(t *testing.T) {
	type decimalTest []string
	var testcases []decimalTest
	raw, err := ioutil.ReadFile("../internal/testdata/decimals.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		d, err := sdk.NewDecFromStr(tc[0])
		require.NoError(t, err)
		r, err := getVRFromGoType(d)
		require.NoError(t, err)
		output, err := r.Format(context.Background(), protoreflect.ValueOf(tc[0]))
		require.NoError(t, err)

		require.Equal(t, []string{tc[1]}, output)
	}
}

func TestGetADR050ValueRenderer(t *testing.T) {
	testcases := []struct {
		name   string
		v      interface{}
		expErr bool
	}{
		{"uint32", uint32(1), false},
		{"uint64", uint64(1), false},
		{"sdk.Int", math.NewInt(1), false},
		{"sdk.Dec", sdk.NewDec(1), false},
		{"float32", float32(1), true},
		{"float64", float64(1), true},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := getVRFromGoType(tc.v)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// getVRFromGoType is like GetADR050ValueRenderer, but taking a Go type
// as input instead of a protoreflect.FieldDescriptor.
func getVRFromGoType(v interface{}) (valuerenderer.ValueRenderer, error) {
	a, b := (&testpb.A{}).ProtoReflect().Descriptor().Fields(), (&testpb.B{}).ProtoReflect().Descriptor().Fields()

	switch v := v.(type) {
	// Valid types for SIGN_MODE_TEXTUAL
	case uint32:
		return valuerenderer.GetADR050ValueRenderer(a.ByName(protoreflect.Name("UINT32")))
	case uint64:
		return valuerenderer.GetADR050ValueRenderer(a.ByName(protoreflect.Name("UINT64")))
	case int32:
		return valuerenderer.GetADR050ValueRenderer(a.ByName(protoreflect.Name("INT32")))
	case int64:
		return valuerenderer.GetADR050ValueRenderer(a.ByName(protoreflect.Name("INT64")))
	case math.Int:
		return valuerenderer.GetADR050ValueRenderer(a.ByName(protoreflect.Name("SDKINT")))
	case sdk.Dec:
		return valuerenderer.GetADR050ValueRenderer(a.ByName(protoreflect.Name("SDKDEC")))

	// Invalid types for SIGN_MODE_TEXTUAL
	case float32:
		return valuerenderer.GetADR050ValueRenderer(b.ByName(protoreflect.Name("FLOAT")))
	case float64:
		return valuerenderer.GetADR050ValueRenderer(b.ByName(protoreflect.Name("FLOAT")))

	default:
		return nil, fmt.Errorf("value %s of type %T not recognized", v, v)
	}
}
