package stake

import (
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// nolint - types is a collection of aliases to the subpackages of this module
type Validator = types.Validator
type Description = types.Description
type Delegation = types.Delegation
type UnbondingDelegation = types.UnbondingDelegation
type Redelegation = types.Redelegation
type Params = types.Params
type Pool = types.Pool
type PoolShares = types.PoolShares
type Keeper = keeper.Keeper
type PrivlegedKeeper = keeper.PrivlegedKeeper
type MsgCreateValidator = types.MsgCreateValidator
type MsgEditValidator = types.MsgEditValidator
type MsgDelegate = types.MsgDelegate
type MsgBeginUnbonding = types.MsgBeginUnbonding
type MsgCompleteUnbonding = types.MsgCompleteUnbonding
type MsgBeginRedelegate = types.MsgBeginRedelegate
type MsgCompleteRedelegate = types.MsgCompleteRedelegate
type GenesisState = types.GenesisState

//function/variable aliases
var (
	NewKeeper                    = keeper.NewKeeper
	NewPrivlegedKeeper           = keeper.NewPrivlegedKeeper
	GetValidatorKey              = keeper.GetValidatorKey
	GetValidatorByPubKeyIndexKey = keeper.GetValidatorByPubKeyIndexKey
	GetValidatorsBondedKey       = keeper.GetValidatorsBondedKey
	GetValidatorsByPowerKey      = keeper.GetValidatorsByPowerKey
	GetTendermintUpdatesKey      = keeper.GetTendermintUpdatesKey
	GetDelegationKey             = keeper.GetDelegationKey
	GetDelegationsKey            = keeper.GetDelegationsKey
	ParamKey                     = keeper.ParamKey
	PoolKey                      = keeper.PoolKey
	ValidatorsKey                = keeper.ValidatorsKey
	ValidatorsByPubKeyIndexKey   = keeper.ValidatorsByPubKeyIndexKey
	ValidatorsBondedKey          = keeper.ValidatorsBondedKey
	ValidatorsByPowerKey         = keeper.ValidatorsByPowerKey
	ValidatorCliffKey            = keeper.ValidatorCliffKey
	ValidatorPowerCliffKey       = keeper.ValidatorPowerCliffKey
	TendermintUpdatesKey         = keeper.TendermintUpdatesKey
	DelegationKey                = keeper.DelegationKey
	IntraTxCounterKey            = keeper.IntraTxCounterKey

	DefaultParams       = types.DefaultParams
	InitialPool         = types.InitialPool
	NewUnbondedShares   = types.NewUnbondedShares
	NewUnbondingShares  = types.NewUnbondingShares
	NewBondedShares     = types.NewBondedShares
	NewValidator        = types.NewValidator
	NewDescription      = types.NewDescription
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState
	RegisterWire        = types.RegisterWire

	// messages
	NewMsgCreateValidator    = types.NewMsgCreateValidator
	NewMsgEditValidator      = types.NewMsgEditValidator
	NewMsgDelegate           = types.NewMsgDelegate
	NewMsgBeginUnbonding     = types.NewMsgBeginUnbonding
	NewMsgCompleteUnbonding  = types.NewMsgCompleteUnbonding
	NewMsgBeginRedelegate    = types.NewMsgBeginRedelegate
	NewMsgCompleteRedelegate = types.NewMsgCompleteRedelegate
)

// errors
const (
	DefaultCodespace     = types.DefaultCodespace
	CodeInvalidValidator = types.CodeInvalidValidator
	CodeInvalidBond      = types.CodeInvalidBond
	CodeInvalidInput     = types.CodeInvalidInput
	CodeValidatorJailed  = types.CodeValidatorJailed
	CodeUnauthorized     = types.CodeUnauthorized
	CodeInternal         = types.CodeInternal
	CodeUnknownRequest   = types.CodeUnknownRequest
)

// nolint
var (
	ErrNotEnoughBondShares   = types.ErrNotEnoughBondShares
	ErrValidatorEmpty        = types.ErrValidatorEmpty
	ErrBadBondingDenom       = types.ErrBadBondingDenom
	ErrBadBondingAmount      = types.ErrBadBondingAmount
	ErrBadSharesPercent      = types.ErrBadSharesPercent
	ErrNoBondingAcct         = types.ErrNoBondingAcct
	ErrCommissionNegative    = types.ErrCommissionNegative
	ErrCommissionHuge        = types.ErrCommissionHuge
	ErrBadValidatorAddr      = types.ErrBadValidatorAddr
	ErrBothShareMsgsGiven    = types.ErrBothShareMsgsGiven
	ErrNeitherShareMsgsGiven = types.ErrNeitherShareMsgsGiven
	ErrBadDelegatorAddr      = types.ErrBadDelegatorAddr
	ErrValidatorExistsAddr   = types.ErrValidatorExistsAddr
	ErrValidatorRevoked      = types.ErrValidatorRevoked
	ErrMissingSignature      = types.ErrMissingSignature
	ErrBondNotNominated      = types.ErrBondNotNominated
	ErrNoValidatorForAddress = types.ErrNoValidatorForAddress
	ErrNoDelegatorForAddress = types.ErrNoDelegatorForAddress
	ErrInsufficientFunds     = types.ErrInsufficientFunds
	ErrBadRemoveValidator    = types.ErrBadRemoveValidator
)
