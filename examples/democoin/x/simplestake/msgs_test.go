package simplestake

import (
	"testing"

	"github.com/stretchr/testify/assert"

	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBondMsgValidation(t *testing.T) {
	privKey := crypto.GenPrivKeyEd25519()
	cases := []struct {
		valid   bool
		msgBond MsgBond
	}{
		{true, NewMsgBond(sdk.Address{}, sdk.NewCoin("mycoin", 5), privKey.PubKey())},
		{false, NewMsgBond(sdk.Address{}, sdk.NewCoin("mycoin", 0), privKey.PubKey())},
	}

	for i, tc := range cases {
		err := tc.msgBond.ValidateBasic()
		if tc.valid {
			assert.Nil(t, err, "%d: %+v", i, err)
		} else {
			assert.NotNil(t, err, "%d", i)
		}
	}
}
