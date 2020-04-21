package helpers

import (
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	"github.com/cosmos/cosmos-sdk/codec"
	"strings"
	"testing"
)

func addFlags(cmd string, flags []string) string {
	for _, f := range flags {
		cmd += " " + f
	}
	return strings.TrimSpace(cmd)
}

//nolint:deadcode,unused
func UnmarshalStdTx(t *testing.T, s string) (stdTx auth.StdTx) {
	cdc := codec.New()
	require.Nil(t, cdc.UnmarshalJSON([]byte(s), &stdTx))
	return
}

func queryEvents(events []string) (out string) {
	for _, event := range events {
		out += event + "&"
	}
	return strings.TrimSuffix(out, "&")
}