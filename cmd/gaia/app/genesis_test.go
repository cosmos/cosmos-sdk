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

		// create the genesis account, give'm few steaks and a buncha token with there name
		accAuth := auth.NewBaseAccountWithAddress(genTx.Address)
		accAuth.Coins = sdk.Coins{
			{genTx.Name + "Token", sdk.NewInt(1000)},
			{"steak", freeFermionsAcc},
		}
		genaccs[i] = NewGenesisAccount(&accAuth)
		stakeData.Pool.LooseTokens = stakeData.Pool.LooseTokens.Add(sdk.NewDecFromInt(freeFermionsAcc)) // increase the supply

		// add the validator
		if len(genTx.Name) > 0 {
			desc := stake.NewDescription(genTx.Name, "", "", "")
			validator := stake.NewValidator(
				sdk.ValAddress(genTx.Address), sdk.MustGetConsPubKeyBech32(genTx.PubKey), desc,
			)

			stakeData.Pool.LooseTokens = stakeData.Pool.LooseTokens.Add(sdk.NewDec(freeFermionVal)) // increase the supply

			// add some new shares to the validator
			var issuedDelShares sdk.Dec
			validator, stakeData.Pool, issuedDelShares = validator.AddTokensFromDel(stakeData.Pool, sdk.NewInt(freeFermionVal))
			stakeData.Validators = append(stakeData.Validators, validator)

			// create the self-delegation from the issuedDelShares
			delegation := stake.Delegation{
				DelegatorAddr: sdk.AccAddress(validator.Operator),
				ValidatorAddr: validator.Operator,
				Shares:        issuedDelShares,
				Height:        0,
			}

			stakeData.Bonds = append(stakeData.Bonds, delegation)
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
