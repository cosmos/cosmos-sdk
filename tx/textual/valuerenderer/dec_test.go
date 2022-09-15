package valuerenderer_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"cosmossdk.io/tx/textual/valuerenderer"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
)

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
