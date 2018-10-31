package app

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/stake"
	stakeTypes "github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var (
	pk1   = ed25519.GenPrivKey().PubKey()
	pk2   = ed25519.GenPrivKey().PubKey()
	pk3   = ed25519.GenPrivKey().PubKey()
	addr1 = sdk.ValAddress(pk1.Address())
	addr2 = sdk.ValAddress(pk2.Address())
	addr3 = sdk.ValAddress(pk3.Address())

	emptyAddr   sdk.ValAddress
	emptyPubkey crypto.PubKey
)

func makeGenesisState(t *testing.T, genTxs []auth.StdTx) GenesisState {
	// start with the default staking genesis state
	stakeData := stake.DefaultGenesisState()
	genAccs := make([]GenesisAccount, len(genTxs))

	for i, genTx := range genTxs {
		msgs := genTx.GetMsgs()
		require.Equal(t, 1, len(msgs))
		msg := msgs[0].(stake.MsgCreateValidator)

		// get genesis flag account information
		genAccs[i] = genesisAccountFromMsgCreateValidator(msg, freeFermionsAcc)
		stakeData.Pool.LooseTokens = stakeData.Pool.LooseTokens.Add(sdk.NewDecFromInt(freeFermionsAcc)) // increase the supply
	}

	// create the final app state
	return GenesisState{
		Accounts:  genAccs,
		StakeData: stakeData,
		GovData:   gov.DefaultGenesisState(),
	}
}

func TestToAccount(t *testing.T) {
	priv := ed25519.GenPrivKey()
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

func makeMsg(name string, pk crypto.PubKey) auth.StdTx {
	desc := stake.NewDescription(name, "", "", "")
	comm := stakeTypes.CommissionMsg{}
	msg := stake.NewMsgCreateValidator(sdk.ValAddress(pk.Address()), pk, sdk.NewInt64Coin("steak", 50), desc, comm)
	return auth.NewStdTx([]sdk.Msg{msg}, auth.StdFee{}, nil, "")
}

func TestGaiaGenesisValidation(t *testing.T) {
	genTxs := make([]auth.StdTx, 2)
	// Test duplicate accounts fails
	genTxs[0] = makeMsg("test-0", pk1)
	genTxs[1] = makeMsg("test-1", pk1)
	genesisState := makeGenesisState(t, genTxs)
	err := GaiaValidateGenesisState(genesisState)
	require.NotNil(t, err)
	// Test bonded + jailed validator fails
	genesisState = makeGenesisState(t, genTxs)
	val1 := stakeTypes.NewValidator(addr1, pk1, stakeTypes.Description{Moniker: "test #2"})
	val1.Jailed = true
	val1.Status = sdk.Bonded
	genesisState.StakeData.Validators = append(genesisState.StakeData.Validators, val1)
	err = GaiaValidateGenesisState(genesisState)
	require.NotNil(t, err)
	// Test duplicate validator fails
	val1.Jailed = false
	genesisState = makeGenesisState(t, genTxs)
	val2 := stakeTypes.NewValidator(addr1, pk1, stakeTypes.Description{Moniker: "test #3"})
	genesisState.StakeData.Validators = append(genesisState.StakeData.Validators, val1)
	genesisState.StakeData.Validators = append(genesisState.StakeData.Validators, val2)
	err = GaiaValidateGenesisState(genesisState)
	require.NotNil(t, err)
}
