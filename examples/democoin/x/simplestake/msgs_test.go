package simplestake

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBondMsgValidation(t *testing.T) {
	privKey := ed25519.GenPrivKey()
	cases := []struct {
		valid   bool
		msgBond MsgBond
	}{
		{true, NewMsgBond(sdk.AccAddress{}, sdk.NewCoin("mycoin", 5), privKey.PubKey())},
		{false, NewMsgBond(sdk.AccAddress{}, sdk.NewCoin("mycoin", 0), privKey.PubKey())},
	}

	for i, tc := range cases {
		err := tc.msgBond.ValidateBasic()
		if tc.valid {
			require.Nil(t, err, "%d: %+v", i, err)
		} else {
			require.NotNil(t, err, "%d", i)
		}
	}
}
