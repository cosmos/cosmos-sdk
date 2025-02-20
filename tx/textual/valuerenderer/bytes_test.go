package valuerenderer_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"

	"cosmossdk.io/tx/textual/valuerenderer"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestBytesJsonTestCases(t *testing.T) {
	var testcases []bytesTest
	// Bytes.json contains bytes that are represented in base64 format, and
	// their expected results in hex.
	raw, err := os.ReadFile("../internal/testdata/bytes.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	textual := valuerenderer.NewTextual(mockCoinMetadataQuerier)

	for _, tc := range testcases {
		data, err := base64.StdEncoding.DecodeString(tc.base64)
		require.NoError(t, err)

		valrend, err := textual.GetValueRenderer(fieldDescriptorFromName("BYTES"))
		require.NoError(t, err)

		screens, err := valrend.Format(context.Background(), protoreflect.ValueOfBytes(data))
		require.NoError(t, err)
		require.Equal(t, 1, len(screens))
		require.Equal(t, tc.hex, screens[0].Text)

		// Round trip
		val, err := valrend.Parse(context.Background(), screens)
		require.NoError(t, err)
		require.Equal(t, tc.base64, base64.StdEncoding.EncodeToString(val.Bytes()))
	}
}

type bytesTest struct {
	hex    string
	base64 string
}

func (t *bytesTest) UnmarshalJSON(b []byte) error {
	a := []interface{}{&t.hex, &t.base64}
	return json.Unmarshal(b, &a)
}
