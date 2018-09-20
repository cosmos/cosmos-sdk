// nolint
package distribution

import (
	"github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/tags"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

type (
	Keeper = keeper.Keeper
	Hooks  = keeper.Hooks

	DelegatorWithdrawInfo = types.DelegatorWithdrawInfo
	DelegatorDistInfo     = types.DelegatorDistInfo
	ValidatorDistInfo     = types.ValidatorDistInfo
	TotalAccum            = types.TotalAccum
	FeePool               = types.FeePool

	MsgSetWithdrawAddress          = types.MsgSetWithdrawAddress
	MsgWithdrawDelegatorRewardsAll = types.MsgWithdrawDelegatorRewardsAll
	MsgWithdrawDelegatorReward     = types.MsgWithdrawDelegatorReward
	MsgWithdrawValidatorRewardsAll = types.MsgWithdrawValidatorRewardsAll

	GenesisState = types.GenesisState
)

var (
	NewKeeper = keeper.NewKeeper

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
	NewMsgWithdrawDelegationReward    = types.NewMsgWithdrawDelegatorReward
	NewMsgWithdrawValidatorRewardsAll = types.NewMsgWithdrawValidatorRewardsAll

	GetCmdWithdrawRewards = cli.GetCmdWithdrawRewards
	GetCmdSetWithdrawAddr = cli.GetCmdSetWithdrawAddr
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
	TagValidator = tags.Validator
	TagDelegator = tags.Delegator
)
