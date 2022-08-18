package valuerenderer_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestFormatBytes(t *testing.T) {
	var testcases []bytesTest
	// Bytes.json contains bytes that are represented in base64 format, and
	// their expected results in hex.
	raw, err := os.ReadFile("../internal/testdata/bytes.json")
	require.NoError(t, err)

	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		data, err := base64.StdEncoding.DecodeString(tc.base64)
		require.NoError(t, err)

		r, err := valueRendererOf(data)
		require.NoError(t, err)

		b := new(strings.Builder)
		err = r.Format(context.Background(), protoreflect.ValueOfBytes(data), b)
		require.NoError(t, err)
		require.Equal(t, tc.hex, b.String())
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
