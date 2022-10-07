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

	textual := valuerenderer.NewTextual(nil)

	for _, tc := range testcases {
		// Parse test case strings as protobuf uint64
		i, err := strconv.ParseUint(tc[0], 10, 64)
		if err == nil {
			r, err := textual.GetValueRenderer(fieldDescriptorFromName("UINT64"))
			require.NoError(t, err)
			b := new(strings.Builder)
			err = r.Format(context.Background(), protoreflect.ValueOf(i), b)
			require.NoError(t, err)

			require.Equal(t, tc[1], b.String())
		}

		// Parse test case strings as protobuf uint32
		i, err = strconv.ParseUint(tc[0], 10, 32)
		if err == nil {
			r, err := textual.GetValueRenderer(fieldDescriptorFromName("UINT32"))
			require.NoError(t, err)
			b := new(strings.Builder)
			err = r.Format(context.Background(), protoreflect.ValueOf(i), b)
			require.NoError(t, err)

			require.Equal(t, tc[1], b.String())
		}

		// Parse test case strings as sdk.Ints
		_, ok := math.NewIntFromString(tc[0])
		if ok {
			r, err := textual.GetValueRenderer(fieldDescriptorFromName("SDKINT"))
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

	textual := valuerenderer.NewTextual(nil)

	for _, tc := range testcases {
		tc := tc
		t.Run(tc[0], func(t *testing.T) {
			r, err := textual.GetValueRenderer(fieldDescriptorFromName("SDKDEC"))
			require.NoError(t, err)
			b := new(strings.Builder)
			err = r.Format(context.Background(), protoreflect.ValueOf(tc[0]), b)
			require.NoError(t, err)

			require.Equal(t, tc[1], b.String())
		})
	}
}

func TestDispatcher(t *testing.T) {
	testcases := []struct {
		name             string
		expErr           bool
		expValueRenderer valuerenderer.ValueRenderer
	}{
		{"UINT32", false, valuerenderer.NewIntValueRenderer()},
		{"UINT64", false, valuerenderer.NewIntValueRenderer()},
		{"SDKINT", false, valuerenderer.NewIntValueRenderer()},
		{"SDKDEC", false, valuerenderer.NewDecValueRenderer()},
		{"BYTES", false, valuerenderer.NewBytesValueRenderer()},
		{"TIMESTAMP", false, valuerenderer.NewTimestampValueRenderer()},
		{"DURATION", false, valuerenderer.NewDurationValueRenderer()},
		{"COIN", false, valuerenderer.NewCoinsValueRenderer(nil)},
		{"COINS", false, valuerenderer.NewCoinsValueRenderer(nil)},
		{"FLOAT", true, nil},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			textual := valuerenderer.NewTextual(nil)
			rend, err := textual.GetValueRenderer(fieldDescriptorFromName(tc.name))

			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.IsType(t, tc.expValueRenderer, rend)
			}
		})
	}
}

// fieldDescriptorFromName is like GetADR050ValueRenderer, but taking a Go type
// as input instead of a protoreflect.FieldDescriptor.
func fieldDescriptorFromName(name string) protoreflect.FieldDescriptor {
	a := (&testpb.A{}).ProtoReflect().Descriptor().Fields()
	fd := a.ByName(protoreflect.Name(name))
	if fd == nil {
		panic(fmt.Errorf("no field descriptor for %s", name))
	}

	return fd
}
