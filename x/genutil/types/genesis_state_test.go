package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/KiraCore/cosmos-sdk/simapp"
	"github.com/KiraCore/cosmos-sdk/simapp/params"
	sdk "github.com/KiraCore/cosmos-sdk/types"
	"github.com/KiraCore/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/KiraCore/cosmos-sdk/x/staking/types"
)

var (
	pk1 = ed25519.GenPrivKey().PubKey()
	pk2 = ed25519.GenPrivKey().PubKey()
)

func TestNetGenesisState(t *testing.T) {
	gen := types.NewGenesisState(nil)
	assert.NotNil(t, gen.GenTxs) // https://github.com/KiraCore/cosmos-sdk/issues/5086

	gen = types.NewGenesisState(
		[]json.RawMessage{
			[]byte(`{"foo":"bar"}`),
		},
	)
	assert.Equal(t, string(gen.GenTxs[0]), `{"foo":"bar"}`)
}

func TestValidateGenesisMultipleMessages(t *testing.T) {
	desc := stakingtypes.NewDescription("testname", "", "", "", "")
	comm := stakingtypes.CommissionRates{}

	msg1 := stakingtypes.NewMsgCreateValidator(sdk.ValAddress(pk1.Address()), pk1,
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 50), desc, comm, sdk.OneInt())

	msg2 := stakingtypes.NewMsgCreateValidator(sdk.ValAddress(pk2.Address()), pk2,
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 50), desc, comm, sdk.OneInt())

	txGen := params.MakeEncodingConfig().TxConfig
	txBuilder := txGen.NewTxBuilder()
	err := txBuilder.SetMsgs(msg1, msg2)
	require.NoError(t, err)

	tx := txBuilder.GetTx()
	genesisState := types.NewGenesisStateFromTx([]sdk.Tx{tx})

	err = types.ValidateGenesis(genesisState, simapp.MakeEncodingConfig().TxConfig.TxJSONDecoder())
	require.Error(t, err)
}

func TestValidateGenesisBadMessage(t *testing.T) {
	desc := stakingtypes.NewDescription("testname", "", "", "", "")

	msg1 := stakingtypes.NewMsgEditValidator(sdk.ValAddress(pk1.Address()), desc, nil, nil)

	txGen := params.MakeEncodingConfig().TxConfig
	txBuilder := txGen.NewTxBuilder()
	err := txBuilder.SetMsgs(msg1)
	require.NoError(t, err)

	tx := txBuilder.GetTx()
	genesisState := types.NewGenesisStateFromTx([]sdk.Tx{tx})

	err = types.ValidateGenesis(genesisState, simapp.MakeEncodingConfig().TxConfig.TxJSONDecoder())
	require.Error(t, err)
}
