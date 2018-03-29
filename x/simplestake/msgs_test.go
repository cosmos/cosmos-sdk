package staking

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
		bondMsg BondMsg
	}{
		{true, NewBondMsg(sdk.Address{}, sdk.Coin{"mycoin", 5}, privKey.PubKey())},
		{false, NewBondMsg(sdk.Address{}, sdk.Coin{"mycoin", 0}, privKey.PubKey())},
	}

	for i, tc := range cases {
		err := tc.bondMsg.ValidateBasic()
		if tc.valid {
			assert.Nil(t, err, "%d: %+v", i, err)
		} else {
			assert.NotNil(t, err, "%d", i)
		}
	}
}
