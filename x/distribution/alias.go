// nolint
package stake

import (
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/querier"
	"github.com/cosmos/cosmos-sdk/x/distribution/tags"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

type (
	Keeper                = keeper.Keeper
	DelegatorWithdrawInfo = types.DelegatorWithdrawInfo
	DelegatorDistInfo     = types.DelegatorDistInfo
	ValidatorDistInfo     = types.ValidatorDistInfo
	TotalAccum            = types.TotalAccum
	FeePool               = types.FeePool

	MsgSetWithdrawAddress          = types.MsgSetWithdrawAddress
	MsgWithdrawDelegatorRewardsAll = types.MsgWithdrawDelegatorRewardsAll
	MsgWithdrawDelegationReward    = types.MsgWithdrawDelegationReward
	MsgWithdrawValidatorRewardsAll = types.MsgWithdrawValidatorRewardsAll

	GenesisState = types.GenesisState
)

var (
	NewKeeper  = keeper.NewKeeper
	NewQuerier = querier.NewQuerier

	GetValidatorDistInfoKey     = keeper.GetValidatorDistInfoKey
	GetDelegationDistInfoKey    = keeper.GetDelegationDistInfoKey
	GetDelegationDistInfosKey   = keeper.GetDelegationDistInfosKey
	GetDelegatorWithdrawAddrKey = keeper.GetDelegatorWithdrawAddrKey
	FeePoolKey                  = keeper.FeePoolKey
	ValidatorDistInfoKey        = keeper.ValidatorDistInfoKey
	DelegatorDistInfoKey        = keeper.DelegatorDistInfoKey
	DelegatorWithdrawInfoKey    = keeper.DelegatorWithdrawInfoKey
	ProposerKey                 = keeper.ProposerKey

	InitialFeePool = types.InitialFeePool

	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState

	RegisterCodec = types.RegisterCodec

	NewMsgSetWithdrawAddress          = types.NewMsgSetWithdrawAddress
	NewMsgWithdrawDelegatorRewardsAll = types.NewMsgWithdrawDelegatorRewardsAll
	NewMsgWithdrawDelegationReward    = types.NewMsgWithdrawDelegationReward
	NewMsgWithdrawValidatorRewardsAll = types.NewMsgWithdrawValidatorRewardsAll
)

const (
	DefaultCodespace = types.DefaultCodespace
	CodeInvalidInput = types.CodeInvalidInput
)

var (
	ErrNilDelegatorAddr = types.ErrNilDelegatorAddr
	ErrNilWithdrawAddr  = types.ErrNilWithdrawAddr
	ErrNilValidatorAddr = types.ErrNilValidatorAddr
)

var (
	ActionModifyWithdrawAddress       = tags.ActionModifyWithdrawAddress
	ActionWithdrawDelegatorRewardsAll = tags.ActionWithdrawDelegatorRewardsAll
	ActionWithdrawDelegatorReward     = tags.ActionWithdrawDelegatorReward
	ActionWithdrawValidatorRewardsAll = tags.ActionWithdrawValidatorRewardsAll

	TagAction    = tags.Action
	TagValidator = tags.SrcValidator
	TagDelegator = tags.Delegator
)
