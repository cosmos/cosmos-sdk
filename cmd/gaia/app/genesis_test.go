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
	addr := pk1.Address()
	// Test duplicate accounts fails
	genTxs[0] = GaiaGenTx{"", sdk.AccAddress(addr), ""}
	genTxs[1] = GaiaGenTx{"", sdk.AccAddress(addr), ""}
	genesisState := makeGenesisState(genTxs)
	err := GaiaValidateGenesisState(genesisState)
	require.NotNil(t, err)
	// Test bonded + revoked validator fails
	genesisState = makeGenesisState(genTxs[:1])
	val1 := stakeTypes.NewValidator(addr1, pk1, stakeTypes.Description{Moniker: "test #2"})
	val1.Jailed = true
	val1.Status = sdk.Bonded
	genesisState.StakeData.Validators = append(genesisState.StakeData.Validators, val1)
	err = GaiaValidateGenesisState(genesisState)
	require.NotNil(t, err)
	// Test duplicate validator fails
	val1.Jailed = false
	genesisState = makeGenesisState(genTxs[:1])
	val2 := stakeTypes.NewValidator(addr1, pk1, stakeTypes.Description{Moniker: "test #3"})
	genesisState.StakeData.Validators = append(genesisState.StakeData.Validators, val1)
	genesisState.StakeData.Validators = append(genesisState.StakeData.Validators, val2)
	err = GaiaValidateGenesisState(genesisState)
	require.NotNil(t, err)
}
