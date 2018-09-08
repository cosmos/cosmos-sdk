package app

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/gov"
	stake "github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func makeGenesisState(genTxs []GaiaGenTx) GenesisState {
	// start with the default staking genesis state
	stakeData := stake.DefaultGenesisState()

	// get genesis flag account information
	genaccs := make([]GenesisAccount, len(genTxs))
	for i, genTx := range genTxs {
		genaccs[i] = genesisAccountFromGenTx(genTx)
		stakeData.Pool.LooseTokens = stakeData.Pool.LooseTokens.Add(sdk.NewDecFromInt(freeFermionsAcc)) // increase the supply

		// add the validator
		if len(genTx.Name) > 0 {
			stakeData = addValidatorToStakeData(genTx, stakeData)
		}
	}

	// create the final app state
	return GenesisState{
		Accounts:  genaccs,
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

func TestGaiaGenesisValidation(t *testing.T) {
	genTxs := make([]GaiaGenTx, 2)
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	// Test duplicate accounts fails
	genTxs[0] = GaiaGenTx{"", sdk.AccAddress(addr), ""}
	genTxs[1] = GaiaGenTx{"", sdk.AccAddress(addr), ""}
	genesisState := makeGenesisState(genTxs)
	err := GaiaValidateGenesisState(genesisState)
	require.NotNil(t, err)
}
