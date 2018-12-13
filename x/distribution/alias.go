// nolint
package distribution

import (
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/tags"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

type (
	Keeper = keeper.Keeper
	Hooks  = keeper.Hooks

	MsgSetWithdrawAddress          = types.MsgSetWithdrawAddress
	MsgWithdrawDelegatorRewardsAll = types.MsgWithdrawDelegatorRewardsAll
	MsgWithdrawDelegatorReward     = types.MsgWithdrawDelegatorReward
	MsgWithdrawValidatorRewardsAll = types.MsgWithdrawValidatorRewardsAll

	GenesisState = types.GenesisState

	// expected keepers
	StakeKeeper         = types.StakeKeeper
	BankKeeper          = types.BankKeeper
	FeeCollectionKeeper = types.FeeCollectionKeeper
)

const (
	DefaultCodespace = types.DefaultCodespace
	CodeInvalidInput = types.CodeInvalidInput
)

var (
	ErrNilDelegatorAddr = types.ErrNilDelegatorAddr
	ErrNilWithdrawAddr  = types.ErrNilWithdrawAddr
	ErrNilValidatorAddr = types.ErrNilValidatorAddr

	TagValidator = tags.Validator
	TagDelegator = tags.Delegator

	NewMsgSetWithdrawAddress          = types.NewMsgSetWithdrawAddress
	NewMsgWithdrawDelegatorReward     = types.NewMsgWithdrawDelegatorReward
	NewMsgWithdrawDelegatorRewardsAll = types.NewMsgWithdrawDelegatorRewardsAll
	NewMsgWithdrawValidatorRewardsAll = types.NewMsgWithdrawValidatorRewardsAll

	NewKeeper         = keeper.NewKeeper
	DefaultParamspace = keeper.DefaultParamspace

	RegisterCodec       = types.RegisterCodec
	DefaultGenesisState = types.DefaultGenesisState
	ValidateGenesis     = types.ValidateGenesis
)
