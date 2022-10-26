package valuerenderer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"cosmossdk.io/tx/textual/valuerenderer"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type stringJsonTest struct {
	Text string
}

func TestStringJsonTestcases(t *testing.T) {
	raw, err := os.ReadFile("../internal/testdata/string.json")
	require.NoError(t, err)

	var testcases []stringJsonTest
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			rend := valuerenderer.NewStringValueRenderer()

			screens, err := rend.Format(context.Background(), protoreflect.ValueOfString(tc.Text))
			require.NoError(t, err)
			require.Equal(t, 1, len(screens))
			require.Equal(t, tc.Text, screens[0].Text)

			val, err := rend.Parse(context.Background(), screens)
			require.NoError(t, err)
			require.Equal(t, tc.Text, val.String())
		})
	}
}

func TestStringHighUnicode(t *testing.T) {
	// We cannot encode Unicode characters beyond the BMP directly in JSON,
	// so this case must be a native Go test.
	s := "\U00101234"
	rend := valuerenderer.NewStringValueRenderer()
	screens, err := rend.Format(context.Background(), protoreflect.ValueOfString(s))
	require.NoError(t, err)
	require.Equal(t, 1, len(screens))
	require.Equal(t, s, screens[0].Text)
	val, err := rend.Parse(context.Background(), screens)
	require.NoError(t, err)
	require.Equal(t, s, val.String())
}
