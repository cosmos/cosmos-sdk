package helpers

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking"
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

func queryEvents(events []string) (out string) {
	for _, event := range events {
		out += event + "&"
	}
	return strings.TrimSuffix(out, "&")
}

//nolint:deadcode,unused
func UnmarshalStdTx(t *testing.T, c *codec.Codec, s string) (stdTx auth.StdTx) {
	require.Nil(t, c.UnmarshalJSON([]byte(s), &stdTx))
	return
}

func FindDelegateAccount(validatorDelegations []staking.Delegation, delegatorAddress string) staking.Delegation {
	for i := 0; i < len(validatorDelegations); i++ {
		if validatorDelegations[i].DelegatorAddress.String() == delegatorAddress {
			return validatorDelegations[i]
		}
	}

	return staking.Delegation{}
}
