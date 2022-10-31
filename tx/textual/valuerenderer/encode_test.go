package valuerenderer_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"cosmossdk.io/tx/textual/valuerenderer"
	"github.com/stretchr/testify/require"
)

func TestEncoding(t *testing.T) {
	for i, tc := range []struct {
		screens  []valuerenderer.Screen
		encoding string
	}{
		{screens: []valuerenderer.Screen{}, encoding: "80"},
		{screens: []valuerenderer.Screen{{}}, encoding: "81a0"},
		{
			screens: []valuerenderer.Screen{
				{Text: "a"}, {Indent: 1}, {Expert: true},
			},
			encoding: "83a1016161a10201a103f5",
		},
		{
			screens: []valuerenderer.Screen{
				{Text: "", Indent: 4, Expert: true},
				{Text: "a", Indent: 0, Expert: true},
				{Text: "b", Indent: 5, Expert: false},
			},
			encoding: "83a2020403f5a201616103f5a20161620205",
		},
		{
			screens: []valuerenderer.Screen{
				{Text: "start"},
				{Text: "middle", Indent: 1},
				{Text: "end"},
			},
			encoding: "83a101657374617274a201666d6964646c650201a10163656e64",
		},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			var buf bytes.Buffer
			err := valuerenderer.Encode(tc.screens, &buf)
			require.NoError(t, err)
			want, err := hex.DecodeString(tc.encoding)
			require.NoError(t, err)
			require.Equal(t, want, buf.Bytes())
		})
	}
}
