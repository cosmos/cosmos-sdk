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
	MsgWithdrawDelegatorReward     = types.MsgWithdrawDelegatorReward
	MsgWithdrawValidatorCommission = types.MsgWithdrawValidatorCommission

	GenesisState = types.GenesisState

	// expected keepers
	StakingKeeper       = types.StakingKeeper
	BankKeeper          = types.BankKeeper
	FeeCollectionKeeper = types.FeeCollectionKeeper

	// querier param types
	QueryValidatorCommissionParams   = keeper.QueryValidatorCommissionParams
	QueryValidatorSlashesParams      = keeper.QueryValidatorSlashesParams
	QueryDelegationRewardsParams     = keeper.QueryDelegationRewardsParams
	QueryDelegatorWithdrawAddrParams = keeper.QueryDelegatorWithdrawAddrParams

	// querier response types
	QueryDelegatorTotalRewardsResponse = types.QueryDelegatorTotalRewardsResponse
	DelegationDelegatorReward          = types.DelegationDelegatorReward
)

const (
	DefaultCodespace = types.DefaultCodespace
	CodeInvalidInput = types.CodeInvalidInput
	StoreKey         = types.StoreKey
	TStoreKey        = types.TStoreKey
	RouterKey        = types.RouterKey
	QuerierRoute     = types.QuerierRoute
)

var (
	ErrNilDelegatorAddr = types.ErrNilDelegatorAddr
	ErrNilWithdrawAddr  = types.ErrNilWithdrawAddr
	ErrNilValidatorAddr = types.ErrNilValidatorAddr

	TagValidator = tags.Validator

	NewMsgSetWithdrawAddress          = types.NewMsgSetWithdrawAddress
	NewMsgWithdrawDelegatorReward     = types.NewMsgWithdrawDelegatorReward
	NewMsgWithdrawValidatorCommission = types.NewMsgWithdrawValidatorCommission

	NewKeeper                                 = keeper.NewKeeper
	NewQuerier                                = keeper.NewQuerier
	NewQueryValidatorOutstandingRewardsParams = keeper.NewQueryValidatorOutstandingRewardsParams
	NewQueryValidatorCommissionParams         = keeper.NewQueryValidatorCommissionParams
	NewQueryValidatorSlashesParams            = keeper.NewQueryValidatorSlashesParams
	NewQueryDelegationRewardsParams           = keeper.NewQueryDelegationRewardsParams
	NewQueryDelegatorParams                   = keeper.NewQueryDelegatorParams
	NewQueryDelegatorWithdrawAddrParams       = keeper.NewQueryDelegatorWithdrawAddrParams
	DefaultParamspace                         = keeper.DefaultParamspace
	RegisterInvariants                        = keeper.RegisterInvariants
	AllInvariants                             = keeper.AllInvariants
	NonNegativeOutstandingInvariant           = keeper.NonNegativeOutstandingInvariant
	CanWithdrawInvariant                      = keeper.CanWithdrawInvariant
	ReferenceCountInvariant                   = keeper.ReferenceCountInvariant
	CreateTestInputDefault                    = keeper.CreateTestInputDefault
	CreateTestInputAdvanced                   = keeper.CreateTestInputAdvanced
	TestAddrs                                 = keeper.TestAddrs

	RegisterCodec       = types.RegisterCodec
	DefaultGenesisState = types.DefaultGenesisState
	ValidateGenesis     = types.ValidateGenesis
	InitialFeePool      = types.InitialFeePool

	// Query types
	QueryParams                      = keeper.QueryParams
	QueryValidatorOutstandingRewards = keeper.QueryValidatorOutstandingRewards
	QueryValidatorCommission         = keeper.QueryValidatorCommission
	QueryValidatorSlashes            = keeper.QueryValidatorSlashes
	QueryDelegationRewards           = keeper.QueryDelegationRewards
	QueryDelegatorTotalRewards       = keeper.QueryDelegatorTotalRewards
	QueryDelegatorValidators         = keeper.QueryDelegatorValidators
	QueryWithdrawAddr                = keeper.QueryWithdrawAddr
	QueryCommunityPool               = keeper.QueryCommunityPool

	// Param types
	ParamCommunityTax        = keeper.ParamCommunityTax
	ParamBaseProposerReward  = keeper.ParamBaseProposerReward
	ParamBonusProposerReward = keeper.ParamBonusProposerReward
	ParamWithdrawAddrEnabled = keeper.ParamWithdrawAddrEnabled
)
