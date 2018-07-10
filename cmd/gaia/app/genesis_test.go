package app

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
)

func TestToAccount(t *testing.T) {
	priv := crypto.GenPrivKeyEd25519()
	addr := sdk.AccAddress(priv.PubKey().Address())
	authAcc := auth.NewBaseAccountWithAddress(addr)
	genAcc := NewGenesisAccount(&authAcc)
	require.Equal(t, authAcc, *genAcc.ToAccount())
}

func TestGaiaAppGenTx(t *testing.T) {
	cdc := MakeCodec()
	_ = cdc

	//TODO test that key overwrite flags work / no overwrites if set off
	//TODO test validator created has provided pubkey
	//TODO test the account created has the correct pubkey
}

func TestGaiaAppGenState(t *testing.T) {
	cdc := MakeCodec()
	_ = cdc

	// TODO test must provide at least genesis transaction
	// TODO test with both one and two genesis transactions:
	// TODO        correct: genesis account created, canididates created, pool token variance
}
