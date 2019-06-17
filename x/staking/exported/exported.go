package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

// DelegationI delegation bond for a delegated proof of stake system
type DelegationI interface {
	GetDelegatorAddr() sdk.AccAddress // delegator sdk.AccAddress for the bond
	GetValidatorAddr() sdk.ValAddress // validator operator address
	GetShares() sdk.Dec               // amount of validator's shares held in this delegation
}

// ValidatorI expected validator functions
type ValidatorI interface {
	IsJailed() bool                                             // whether the validator is jailed
	GetMoniker() string                                         // moniker of the validator
	GetStatus() sdk.BondStatus                                  // status of the validator
	IsBonded() bool                                             // check if has a bonded status
	IsUnbonded() bool                                           // check if has status unbonded
	IsUnbonding() bool                                          // check if has status unbonding
	GetOperator() sdk.ValAddress                                // operator address to receive/return validators coins
	GetConsPubKey() crypto.PubKey                               // validation consensus pubkey
	GetConsAddr() sdk.ConsAddress                               // validation consensus address
	GetTokens() sdk.Int                                         // validation tokens
	GetBondedTokens() sdk.Int                                   // validator bonded tokens
	GetConsensusPower() int64                                   // validation power in tendermint
	GetCommission() sdk.Dec                                     // validator commission rate
	GetMinSelfDelegation() sdk.Int                              // validator minimum self delegation
	GetDelegatorShares() sdk.Dec                                // total outstanding delegator shares
	TokensFromShares(sdk.Dec) sdk.Dec                           // token worth of provided delegator shares
	TokensFromSharesTruncated(sdk.Dec) sdk.Dec                  // token worth of provided delegator shares, truncated
	TokensFromSharesRoundUp(sdk.Dec) sdk.Dec                    // token worth of provided delegator shares, rounded up
	SharesFromTokens(amt sdk.Int) (sdk.Dec, sdk.Error)          // shares worth of delegator's bond
	SharesFromTokensTruncated(amt sdk.Int) (sdk.Dec, sdk.Error) // truncated shares worth of delegator's bond
}
