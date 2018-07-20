// nolint
package stake

import (
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/tags"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

type (
	Keeper                = keeper.Keeper
	Validator             = types.Validator
	BechValidator         = types.BechValidator
	Description           = types.Description
	Delegation            = types.Delegation
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
)

var (
	NewKeeper = keeper.NewKeeper

	GetValidatorKey              = keeper.GetValidatorKey
	GetValidatorByPubKeyIndexKey = keeper.GetValidatorByPubKeyIndexKey
	GetValidatorsBondedIndexKey  = keeper.GetValidatorsBondedIndexKey
	GetValidatorsByPowerIndexKey = keeper.GetValidatorsByPowerIndexKey
	GetTendermintUpdatesKey      = keeper.GetTendermintUpdatesKey
	GetDelegationKey             = keeper.GetDelegationKey
	GetDelegationsKey            = keeper.GetDelegationsKey
	ParamKey                     = keeper.ParamKey
	PoolKey                      = keeper.PoolKey
	ValidatorsKey                = keeper.ValidatorsKey
	ValidatorsByPubKeyIndexKey   = keeper.ValidatorsByPubKeyIndexKey
	ValidatorsBondedIndexKey     = keeper.ValidatorsBondedIndexKey
	ValidatorsByPowerIndexKey    = keeper.ValidatorsByPowerIndexKey
	ValidatorCliffIndexKey       = keeper.ValidatorCliffIndexKey
	ValidatorPowerCliffKey       = keeper.ValidatorPowerCliffKey
	TendermintUpdatesKey         = keeper.TendermintUpdatesKey
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

	DefaultParams       = types.DefaultParams
	InitialPool         = types.InitialPool
	NewValidator        = types.NewValidator
	NewDescription      = types.NewDescription
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState
	RegisterWire        = types.RegisterWire

	NewMsgCreateValidator           = types.NewMsgCreateValidator
	NewMsgCreateValidatorOnBehalfOf = types.NewMsgCreateValidatorOnBehalfOf
	NewMsgEditValidator             = types.NewMsgEditValidator
	NewMsgDelegate                  = types.NewMsgDelegate
	NewMsgBeginUnbonding            = types.NewMsgBeginUnbonding
	NewMsgCompleteUnbonding         = types.NewMsgCompleteUnbonding
	NewMsgBeginRedelegate           = types.NewMsgBeginRedelegate
	NewMsgCompleteRedelegate        = types.NewMsgCompleteRedelegate
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
	ErrValidatorRevoked      = types.ErrValidatorRevoked
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
