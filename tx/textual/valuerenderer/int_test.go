package valuerenderer_test

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/math"
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
