package genutil

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var (
	pk1 = ed25519.GenPrivKey().PubKey()
	pk2 = ed25519.GenPrivKey().PubKey()
)

func TestValidateGenesisMultipleMessages(t *testing.T) {

	desc := staking.NewDescription("testname", "", "", "")
	comm := staking.CommissionRates{}

	msg1 := staking.NewMsgCreateValidator(sdk.ValAddress(pk1.Address()), pk1,
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 50), desc, comm, sdk.OneInt())

	msg2 := staking.NewMsgCreateValidator(sdk.ValAddress(pk2.Address()), pk2,
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 50), desc, comm, sdk.OneInt())

	genTxs := auth.NewStdTx([]sdk.Msg{msg1, msg2}, auth.StdFee{}, nil, "")
	genesisState := NewGenesisStateFromStdTx([]auth.StdTx{genTxs})

	err := ValidateGenesis(genesisState)
	require.Error(t, err)
}

func TestValidateGenesisBadMessage(t *testing.T) {
	desc := staking.NewDescription("testname", "", "", "")

	msg1 := staking.NewMsgEditValidator(sdk.ValAddress(pk1.Address()), desc, nil, nil)

	genTxs := auth.NewStdTx([]sdk.Msg{msg1}, auth.StdFee{}, nil, "")
	genesisState := NewGenesisStateFromStdTx([]auth.StdTx{genTxs})

	err := ValidateGenesis(genesisState)
	require.Error(t, err)
}
