package types

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	pk1 = ed25519.GenPrivKey().PubKey()
	pk2 = ed25519.GenPrivKey().PubKey()
)

func TestValidateGenesisMultipleMessages(t *testing.T) {

	desc := stakingtypes.NewDescription("testname", "", "", "")
	comm := stakingtypes.CommissionRates{}

	msg1 := stakingtypes.NewMsgCreateValidator(sdk.ValAddress(pk1.Address()), pk1,
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 50), desc, comm, sdk.OneInt())

	msg2 := stakingtypes.NewMsgCreateValidator(sdk.ValAddress(pk2.Address()), pk2,
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 50), desc, comm, sdk.OneInt())

	genTxs := authtypes.NewStdTx([]sdk.Msg{msg1, msg2}, authtypes.StdFee{}, nil, "")
	genesisState := NewGenesisStateFromStdTx([]authtypes.StdTx{genTxs})

	err := ValidateGenesis(genesisState)
	require.Error(t, err)
}

func TestValidateGenesisBadMessage(t *testing.T) {
	desc := stakingtypes.NewDescription("testname", "", "", "")

	msg1 := stakingtypes.NewMsgEditValidator(sdk.ValAddress(pk1.Address()), desc, nil, nil)

	genTxs := authtypes.NewStdTx([]sdk.Msg{msg1}, authtypes.StdFee{}, nil, "")
	genesisState := NewGenesisStateFromStdTx([]authtypes.StdTx{genTxs})

	err := ValidateGenesis(genesisState)
	require.Error(t, err)
}
