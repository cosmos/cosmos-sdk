package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// GenStakingGenesisState generates a random GenesisState for staking
func GenStakingGenesisState(
	cdc *codec.Codec, r *rand.Rand, accs []simulation.Account, amount, numAccs, numInitiallyBonded int64,
	ap simulation.AppParams, genesisState map[string]json.RawMessage,
) staking.GenesisState {

	stakingGenesis := staking.NewGenesisState(
		staking.NewParams(
			func(r *rand.Rand) time.Duration {
				var v time.Duration
				ap.GetOrGenerate(cdc, simulation.UnbondingTime, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.UnbondingTime](r).(time.Duration)
					})
				return v
			}(r),
			func(r *rand.Rand) uint16 {
				var v uint16
				ap.GetOrGenerate(cdc, simulation.MaxValidators, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.MaxValidators](r).(uint16)
					})
				return v
			}(r),
			7,
			sdk.DefaultBondDenom,
		),
		nil,
		nil,
	)

	var (
		validators  []staking.Validator
		delegations []staking.Delegation
	)

	valAddrs := make([]sdk.ValAddress, numInitiallyBonded)
	for i := 0; i < int(numInitiallyBonded); i++ {
		valAddr := sdk.ValAddress(accs[i].Address)
		valAddrs[i] = valAddr

		validator := staking.NewValidator(valAddr, accs[i].PubKey, staking.Description{})
		validator.Tokens = sdk.NewInt(amount)
		validator.DelegatorShares = sdk.NewDec(amount)
		delegation := staking.NewDelegation(accs[i].Address, valAddr, sdk.NewDec(amount))
		validators = append(validators, validator)
		delegations = append(delegations, delegation)
	}

	stakingGenesis.Validators = validators
	stakingGenesis.Delegations = delegations

	fmt.Printf("Selected randomly generated staking parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, stakingGenesis.Params))
	genesisState[staking.ModuleName] = cdc.MustMarshalJSON(stakingGenesis)

	return stakingGenesis
}
