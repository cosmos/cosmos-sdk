package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// expected bank keeper
type BankKeeper interface {
	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

	// TODO remove once governance doesn't require use of accounts
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
	SetSendEnabled(ctx sdk.Context, enabled bool)
}

// StakingKeeper expected staking keeper (Validator and Delegator sets)
type StakingKeeper interface {
	// iterate through bonded validators by operator address, execute func for each validator
	IterateBondedValidatorsByPower(sdk.Context,
		func(index int64, validator exported.ValidatorI) (stop bool))

	TotalBondedTokens(sdk.Context) sdk.Int // total bonded tokens within the validator set

	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress,
		fn func(index int64, delegation exported.DelegationI) (stop bool))
}
