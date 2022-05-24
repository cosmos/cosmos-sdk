package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	pk1 = ed25519.GenPrivKey().PubKey()
	pk2 = ed25519.GenPrivKey().PubKey()
)

func TestNetGenesisState(t *testing.T) {
	gen := types.NewGenesisState(nil)
	assert.NotNil(t, gen.GenTxs) // https://github.com/cosmos/cosmos-sdk/issues/5086

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

	msg1, err := stakingtypes.NewMsgCreateValidator(sdk.ValAddress(pk1.Address()), pk1,
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 50), desc, comm, sdk.OneInt())
	require.NoError(t, err)

	msg2, err := stakingtypes.NewMsgCreateValidator(sdk.ValAddress(pk2.Address()), pk2,
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 50), desc, comm, sdk.OneInt())
	require.NoError(t, err)

	txGen := simapp.MakeTestEncodingConfig().TxConfig
	txBuilder := txGen.NewTxBuilder()
	require.NoError(t, txBuilder.SetMsgs(msg1, msg2))

	tx := txBuilder.GetTx()
	genesisState := types.NewGenesisStateFromTx(txGen.TxJSONEncoder(), []sdk.Tx{tx})

	err = types.ValidateGenesis(genesisState, simapp.MakeTestEncodingConfig().TxConfig.TxJSONDecoder())
	require.Error(t, err)
}

func TestValidateGenesisBadMessage(t *testing.T) {
	desc := stakingtypes.NewDescription("testname", "", "", "", "")

	msg1 := stakingtypes.NewMsgEditValidator(sdk.ValAddress(pk1.Address()), desc, nil, nil)

	txGen := simapp.MakeTestEncodingConfig().TxConfig
	txBuilder := txGen.NewTxBuilder()
	err := txBuilder.SetMsgs(msg1)
	require.NoError(t, err)

	tx := txBuilder.GetTx()
	genesisState := types.NewGenesisStateFromTx(txGen.TxJSONEncoder(), []sdk.Tx{tx})

	err = types.ValidateGenesis(genesisState, simapp.MakeTestEncodingConfig().TxConfig.TxJSONDecoder())
	require.Error(t, err)
}

func TestGenesisStateFromGenFile(t *testing.T) {
	cdc := codec.NewLegacyAmino()

	genFile := "../../../tests/fixtures/adr-024-coin-metadata_genesis.json"
	genesisState, _, err := types.GenesisStateFromGenFile(genFile)
	require.NoError(t, err)

	var bankGenesis banktypes.GenesisState
	cdc.MustUnmarshalJSON(genesisState[banktypes.ModuleName], &bankGenesis)

	require.True(t, bankGenesis.Params.DefaultSendEnabled)
	require.Equal(t, "1000nametoken,100000000stake", bankGenesis.Balances[0].GetCoins().String())
	require.Equal(t, "cosmos106vrzv5xkheqhjm023pxcxlqmcjvuhtfyachz4", bankGenesis.Balances[0].GetAddress().String())
	require.Equal(t, "The native staking token of the Cosmos Hub.", bankGenesis.DenomMetadata[0].GetDescription())
	require.Equal(t, "uatom", bankGenesis.DenomMetadata[0].GetBase())
	require.Equal(t, "matom", bankGenesis.DenomMetadata[0].GetDenomUnits()[1].GetDenom())
	require.Equal(t, []string{"milliatom"}, bankGenesis.DenomMetadata[0].GetDenomUnits()[1].GetAliases())
	require.Equal(t, uint32(3), bankGenesis.DenomMetadata[0].GetDenomUnits()[1].GetExponent())

}
