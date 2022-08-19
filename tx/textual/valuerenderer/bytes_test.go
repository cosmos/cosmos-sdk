package valuerenderer_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestFormatBytes(t *testing.T) {
	var testcases []bytesTest
	raw, err := os.ReadFile("../internal/testdata/bytes.json")
	require.NoError(t, err)

	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		data, err := hex.DecodeString(tc.hex)
		require.NoError(t, err)

		r, err := valueRendererOf(data)
		require.NoError(t, err)

		b := new(strings.Builder)
		err = r.Format(context.Background(), protoreflect.ValueOfBytes(data), b)
		require.NoError(t, err)
		require.Equal(t, tc.expRes, b.String())
	}
}

type bytesTest struct {
	hex    string
	expRes string
}

func (t *bytesTest) UnmarshalJSON(b []byte) error {
	a := []interface{}{&t.hex, &t.expRes}
	return json.Unmarshal(b, &a)
}
