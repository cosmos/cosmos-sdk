// nolint
package stake

import (
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/querier"
	"github.com/cosmos/cosmos-sdk/x/stake/tags"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

type (
	Keeper                = keeper.Keeper
	Validator             = types.Validator
	Description           = types.Description
	Commission            = types.Commission
	Delegation            = types.Delegation
	DelegationSummary     = types.DelegationSummary
	UnbondingDelegation   = types.UnbondingDelegation
	Redelegation          = types.Redelegation
	Params                = types.Params
	Pool                  = types.Pool
	MsgCreateValidator    = types.MsgCreateValidator
	MsgEditValidator      = types.MsgEditValidator
	MsgDelegate           = types.MsgDelegate
	MsgBeginUnbonding     = types.MsgBeginUnbonding
	MsgCompleteUnbonding  = types.MsgCompleteUnbonding
	MsgBeginRedelegate    = types.MsgBeginRedelegate
	MsgCompleteRedelegate = types.MsgCompleteRedelegate
	GenesisState          = types.GenesisState
	QueryDelegatorParams  = querier.QueryDelegatorParams
	QueryValidatorParams  = querier.QueryValidatorParams
	QueryBondsParams      = querier.QueryBondsParams
)

var (
	NewKeeper = keeper.NewKeeper

	GetValidatorKey              = keeper.GetValidatorKey
	GetValidatorByConsAddrKey    = keeper.GetValidatorByConsAddrKey
	GetValidatorsByPowerIndexKey = keeper.GetValidatorsByPowerIndexKey
	GetDelegationKey             = keeper.GetDelegationKey
	GetDelegationsKey            = keeper.GetDelegationsKey
	PoolKey                      = keeper.PoolKey
	ValidatorsKey                = keeper.ValidatorsKey
	ValidatorsByConsAddrKey      = keeper.ValidatorsByConsAddrKey
	ValidatorsBondedIndexKey     = keeper.ValidatorsBondedIndexKey
	ValidatorsByPowerIndexKey    = keeper.ValidatorsByPowerIndexKey
	DelegationKey                = keeper.DelegationKey
	IntraTxCounterKey            = keeper.IntraTxCounterKey
	GetUBDKey                    = keeper.GetUBDKey
	GetUBDByValIndexKey          = keeper.GetUBDByValIndexKey
	GetUBDsKey                   = keeper.GetUBDsKey
	GetUBDsByValIndexKey         = keeper.GetUBDsByValIndexKey
	GetREDKey                    = keeper.GetREDKey
	GetREDByValSrcIndexKey       = keeper.GetREDByValSrcIndexKey
	GetREDByValDstIndexKey       = keeper.GetREDByValDstIndexKey
	GetREDsKey                   = keeper.GetREDsKey
	GetREDsFromValSrcIndexKey    = keeper.GetREDsFromValSrcIndexKey
	GetREDsToValDstIndexKey      = keeper.GetREDsToValDstIndexKey
	GetREDsByDelToValDstIndexKey = keeper.GetREDsByDelToValDstIndexKey

	DefaultParamspace      = keeper.DefaultParamspace
	KeyInflationRateChange = types.KeyInflationRateChange
	KeyInflationMax        = types.KeyInflationMax
	KeyGoalBonded          = types.KeyGoalBonded
	KeyUnbondingTime       = types.KeyUnbondingTime
	KeyMaxValidators       = types.KeyMaxValidators
	KeyBondDenom           = types.KeyBondDenom

	DefaultParams         = types.DefaultParams
	InitialPool           = types.InitialPool
	NewValidator          = types.NewValidator
	NewDescription        = types.NewDescription
	NewCommission         = types.NewCommission
	NewCommissionMsg      = types.NewCommissionMsg
	NewCommissionWithTime = types.NewCommissionWithTime
	NewGenesisState       = types.NewGenesisState
	DefaultGenesisState   = types.DefaultGenesisState
	RegisterCodec         = types.RegisterCodec

	NewMsgCreateValidator           = types.NewMsgCreateValidator
	NewMsgCreateValidatorOnBehalfOf = types.NewMsgCreateValidatorOnBehalfOf
	NewMsgEditValidator             = types.NewMsgEditValidator
	NewMsgDelegate                  = types.NewMsgDelegate
	NewMsgBeginUnbonding            = types.NewMsgBeginUnbonding
	NewMsgCompleteUnbonding         = types.NewMsgCompleteUnbonding
	NewMsgBeginRedelegate           = types.NewMsgBeginRedelegate
	NewMsgCompleteRedelegate        = types.NewMsgCompleteRedelegate

	NewQuerier = querier.NewQuerier
)

const (
	QueryValidators          = querier.QueryValidators
	QueryValidator           = querier.QueryValidator
	QueryDelegator           = querier.QueryDelegator
	QueryDelegation          = querier.QueryDelegation
	QueryUnbondingDelegation = querier.QueryUnbondingDelegation
	QueryDelegatorValidators = querier.QueryDelegatorValidators
	QueryDelegatorValidator  = querier.QueryDelegatorValidator
	QueryPool                = querier.QueryPool
	QueryParameters          = querier.QueryParameters
)

const (
	DefaultCodespace      = types.DefaultCodespace
	CodeInvalidValidator  = types.CodeInvalidValidator
	CodeInvalidDelegation = types.CodeInvalidDelegation
	CodeInvalidInput      = types.CodeInvalidInput
	CodeValidatorJailed   = types.CodeValidatorJailed
	CodeUnauthorized      = types.CodeUnauthorized
	CodeInternal          = types.CodeInternal
	CodeUnknownRequest    = types.CodeUnknownRequest
)

var (
	ErrNilValidatorAddr      = types.ErrNilValidatorAddr
	ErrNoValidatorFound      = types.ErrNoValidatorFound
	ErrValidatorOwnerExists  = types.ErrValidatorOwnerExists
	ErrValidatorPubKeyExists = types.ErrValidatorPubKeyExists
	ErrValidatorJailed       = types.ErrValidatorJailed
	ErrBadRemoveValidator    = types.ErrBadRemoveValidator
	ErrDescriptionLength     = types.ErrDescriptionLength
	ErrCommissionNegative    = types.ErrCommissionNegative
	ErrCommissionHuge        = types.ErrCommissionHuge

	ErrNilDelegatorAddr          = types.ErrNilDelegatorAddr
	ErrBadDenom                  = types.ErrBadDenom
	ErrBadDelegationAmount       = types.ErrBadDelegationAmount
	ErrNoDelegation              = types.ErrNoDelegation
	ErrBadDelegatorAddr          = types.ErrBadDelegatorAddr
	ErrNoDelegatorForAddress     = types.ErrNoDelegatorForAddress
	ErrInsufficientShares        = types.ErrInsufficientShares
	ErrDelegationValidatorEmpty  = types.ErrDelegationValidatorEmpty
	ErrNotEnoughDelegationShares = types.ErrNotEnoughDelegationShares
	ErrBadSharesAmount           = types.ErrBadSharesAmount
	ErrBadSharesPercent          = types.ErrBadSharesPercent

	ErrNotMature             = types.ErrNotMature
	ErrNoUnbondingDelegation = types.ErrNoUnbondingDelegation
	ErrNoRedelegation        = types.ErrNoRedelegation
	ErrBadRedelegationDst    = types.ErrBadRedelegationDst

	ErrBothShareMsgsGiven    = types.ErrBothShareMsgsGiven
	ErrNeitherShareMsgsGiven = types.ErrNeitherShareMsgsGiven
	ErrMissingSignature      = types.ErrMissingSignature
)

var (
	ActionCreateValidator      = tags.ActionCreateValidator
	ActionEditValidator        = tags.ActionEditValidator
	ActionDelegate             = tags.ActionDelegate
	ActionBeginUnbonding       = tags.ActionBeginUnbonding
	ActionCompleteUnbonding    = tags.ActionCompleteUnbonding
	ActionBeginRedelegation    = tags.ActionBeginRedelegation
	ActionCompleteRedelegation = tags.ActionCompleteRedelegation

	TagAction       = tags.Action
	TagSrcValidator = tags.SrcValidator
	TagDstValidator = tags.DstValidator
	TagDelegator    = tags.Delegator
	TagMoniker      = tags.Moniker
	TagIdentity     = tags.Identity
)

// nolint - reexport
func ParamTable() params.Table {
	return keeper.ParamTable()
}
