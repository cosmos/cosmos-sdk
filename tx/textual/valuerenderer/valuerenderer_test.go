package valuerenderer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	tspb "google.golang.org/protobuf/types/known/timestamppb"

	"cosmossdk.io/math"
	"cosmossdk.io/tx/textual/internal/testpb"
	"cosmossdk.io/tx/textual/valuerenderer"
)

func TestFormatInteger(t *testing.T) {
	type integerTest []string
	var testcases []integerTest
	raw, err := os.ReadFile("../internal/testdata/integers.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		// Parse test case strings as protobuf uint64
		i, err := strconv.ParseUint(tc[0], 10, 64)
		if err == nil {
			r, err := valueRendererOf(i)
			require.NoError(t, err)
			b := new(strings.Builder)
			err = r.Format(context.Background(), protoreflect.ValueOf(i), b)
			require.NoError(t, err)

			require.Equal(t, tc[1], b.String())
		}

		// Parse test case strings as protobuf uint32
		i, err = strconv.ParseUint(tc[0], 10, 32)
		if err == nil {
			r, err := valueRendererOf(i)
			require.NoError(t, err)
			b := new(strings.Builder)
			err = r.Format(context.Background(), protoreflect.ValueOf(i), b)
			require.NoError(t, err)

			require.Equal(t, tc[1], b.String())
		}

		// Parse test case strings as sdk.Ints
		sdkInt, ok := math.NewIntFromString(tc[0])
		if ok {
			r, err := valueRendererOf(sdkInt)
			require.NoError(t, err)
			b := new(strings.Builder)
			err = r.Format(context.Background(), protoreflect.ValueOf(tc[0]), b)
			require.NoError(t, err)

			require.Equal(t, tc[1], b.String())
		}
	}
}

func TestFormatDecimal(t *testing.T) {
	type decimalTest []string
	var testcases []decimalTest
	raw, err := os.ReadFile("../internal/testdata/decimals.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		tc := tc
		t.Run(tc[0], func(t *testing.T) {
			d, err := math.LegacyNewDecFromStr(tc[0])
			require.NoError(t, err)
			r, err := valueRendererOf(d)
			require.NoError(t, err)
			b := new(strings.Builder)
			err = r.Format(context.Background(), protoreflect.ValueOf(tc[0]), b)
			require.NoError(t, err)

			require.Equal(t, tc[1], b.String())
		})
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
		{"sdk.Dec", math.LegacyNewDec(1), false},
		{"[]byte", []byte{1}, false},
		{"float32", float32(1), true},
		{"float64", float64(1), true},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := valueRendererOf(tc.v)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTimestampDispatch(t *testing.T) {
	a := (&testpb.A{}).ProtoReflect().Descriptor().Fields()
	textual := valuerenderer.NewTextual()
	rend, err := textual.GetValueRenderer(a.ByName(protoreflect.Name("TIMESTAMP")))
	require.NoError(t, err)
	require.IsType(t, valuerenderer.NewTimestampValueRenderer(), rend)
}

// valueRendererOf is like GetADR050ValueRenderer, but taking a Go type
// as input instead of a protoreflect.FieldDescriptor.
func valueRendererOf(v interface{}) (valuerenderer.ValueRenderer, error) {
	a, b := (&testpb.A{}).ProtoReflect().Descriptor().Fields(), (&testpb.B{}).ProtoReflect().Descriptor().Fields()

	textual := valuerenderer.NewTextual()
	switch v := v.(type) {
	// Valid types for SIGN_MODE_TEXTUAL
	case uint32:
		return textual.GetValueRenderer(a.ByName(protoreflect.Name("UINT32")))
	case uint64:
		return textual.GetValueRenderer(a.ByName(protoreflect.Name("UINT64")))
	case int32:
		return textual.GetValueRenderer(a.ByName(protoreflect.Name("INT32")))
	case int64:
		return textual.GetValueRenderer(a.ByName(protoreflect.Name("INT64")))
	case []byte:
		return textual.GetValueRenderer(a.ByName(protoreflect.Name("BYTES")))
	case math.Int:
		return textual.GetValueRenderer(a.ByName(protoreflect.Name("SDKINT")))
	case math.LegacyDec:
		return textual.GetValueRenderer(a.ByName(protoreflect.Name("SDKDEC")))
	case tspb.Timestamp:
		return textual.GetValueRenderer(a.ByName(protoreflect.Name("TIMESTAMP")))

	// Invalid types for SIGN_MODE_TEXTUAL
	case float32:
		return textual.GetValueRenderer(b.ByName(protoreflect.Name("FLOAT")))
	case float64:
		return textual.GetValueRenderer(b.ByName(protoreflect.Name("FLOAT")))

	default:
		return nil, fmt.Errorf("value %s of type %T not recognized", v, v)
	}
}
