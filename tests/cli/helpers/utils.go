package helpers

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func AddFlags(cmd string, flags []string) string {
	for _, f := range flags {
		cmd += " " + f
	}
	return strings.TrimSpace(cmd)
}

//nolint:deadcode,unused
func UnmarshalStdTx(t *testing.T, c *codec.Codec, s string) (stdTx auth.StdTx) {
	require.Nil(t, c.UnmarshalJSON([]byte(s), &stdTx))
	return
}
