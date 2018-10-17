package app

import (
	"encoding/json"
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/stake"
	tmtypes "github.com/tendermint/tendermint/types"
)

// NewTestGaiaAppGenState creates the core parameters for a test genesis
// initialization given a set of genesis txs, TM validators and their respective
// operating addresses.
func NewTestGaiaAppGenState(
	cdc *codec.Codec, appGenTxs []json.RawMessage, tmVals []tmtypes.GenesisValidator, valOperAddrs []sdk.ValAddress,
) (GenesisState, error) {

	switch {
	case len(appGenTxs) == 0:
		return GenesisState{}, errors.New("must provide at least genesis transaction")
	case len(tmVals) != len(valOperAddrs):
		return GenesisState{}, errors.New("number of TM validators does not match number of operator addresses")
	}

	// start with the default staking genesis state
	stakeData := stake.DefaultGenesisState()

	// get genesis account information
	genAccs := make([]GenesisAccount, len(appGenTxs))
	for i, appGenTx := range appGenTxs {

		var genTx GaiaGenTx
		if err := cdc.UnmarshalJSON(appGenTx, &genTx); err != nil {
			return GenesisState{}, err
		}

		stakeData.Pool.LooseTokens = stakeData.Pool.LooseTokens.Add(sdk.NewDecFromInt(freeFermionsAcc))

		// create the genesis account for the given genesis tx
		genAccs[i] = genesisAccountFromGenTx(genTx)
	}

	for i, tmVal := range tmVals {
		var issuedDelShares sdk.Dec

		// increase total supply by validator's power
		power := sdk.NewInt(tmVal.Power)
		stakeData.Pool.LooseTokens = stakeData.Pool.LooseTokens.Add(sdk.NewDecFromInt(power))

		// add the validator
		desc := stake.NewDescription(tmVal.Name, "", "", "")
		validator := stake.NewValidator(valOperAddrs[i], tmVal.PubKey, desc)

		validator, stakeData.Pool, issuedDelShares = validator.AddTokensFromDel(stakeData.Pool, power)
		stakeData.Validators = append(stakeData.Validators, validator)

		// create the self-delegation from the issuedDelShares
		selfDel := stake.Delegation{
			DelegatorAddr: sdk.AccAddress(validator.OperatorAddr),
			ValidatorAddr: validator.OperatorAddr,
			Shares:        issuedDelShares,
			Height:        0,
		}

		stakeData.Bonds = append(stakeData.Bonds, selfDel)
	}

	return GenesisState{
		Accounts:     genAccs,
		StakeData:    stakeData,
		DistrData:    distr.DefaultGenesisState(),
		SlashingData: slashing.DefaultGenesisState(),
		GovData:      gov.DefaultGenesisState(),
	}, nil
}
