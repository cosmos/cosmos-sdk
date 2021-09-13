package cosmovisor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureChecksumUrl(t *testing.T) {
	cases := map[string]struct {
		url string
		err string
	}{
		"no checksum": {
			url: "http://abc.xyz/a/b?aa=1",
			err: "checksum must be included",
		},
	}

	// TODO add more tests

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := ensureChecksumUrl(tc.url)
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.err)
			}
		})
	}
}
