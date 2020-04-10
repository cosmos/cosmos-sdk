package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingexported "github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// StakingKeeper expected staking keeper (Validator and Delegator sets)
type StakingKeeper interface {
	// iterate through bonded validators by operator address, execute func for each validator
	IterateBondedValidatorsByPower(sdk.Context,
		func(index int64, validator stakingexported.ValidatorI) (stop bool))

	TotalBondedTokens(sdk.Context) sdk.Int // total bonded tokens within the validator set
}
