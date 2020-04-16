package bank

// nolint

import (
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	QueryBalance       = types.QueryBalance
	QueryAllBalances   = types.QueryAllBalances
	ModuleName         = types.ModuleName
	RouterKey          = types.RouterKey
	StoreKey           = types.StoreKey
	DefaultParamspace  = types.DefaultParamspace
	DefaultSendEnabled = types.DefaultSendEnabled

	EventTypeTransfer      = types.EventTypeTransfer
	AttributeKeyRecipient  = types.AttributeKeyRecipient
	AttributeKeySender     = types.AttributeKeySender
	AttributeValueCategory = types.AttributeValueCategory
)

var (
	RegisterInvariants          = keeper.RegisterInvariants
	NonnegativeBalanceInvariant = keeper.NonnegativeBalanceInvariant
	NewBaseKeeper               = keeper.NewBaseKeeper
	NewBaseSendKeeper           = keeper.NewBaseSendKeeper
	NewBaseViewKeeper           = keeper.NewBaseViewKeeper
	RegisterCodec               = types.RegisterCodec
	ErrNoInputs                 = types.ErrNoInputs
	ErrNoOutputs                = types.ErrNoOutputs
	ErrInputOutputMismatch      = types.ErrInputOutputMismatch
	ErrSendDisabled             = types.ErrSendDisabled
	NewGenesisState             = types.NewGenesisState
	DefaultGenesisState         = types.DefaultGenesisState
	ValidateGenesis             = types.ValidateGenesis
	SanitizeGenesisBalances     = types.SanitizeGenesisBalances
	GetGenesisStateFromAppState = types.GetGenesisStateFromAppState
	NewMsgSend                  = types.NewMsgSend
	NewMsgMultiSend             = types.NewMsgMultiSend
	NewInput                    = types.NewInput
	NewOutput                   = types.NewOutput
	ValidateInputsOutputs       = types.ValidateInputsOutputs
	ParamKeyTable               = types.ParamKeyTable
	NewQueryBalanceParams       = types.NewQueryBalanceRequest
	NewQueryAllBalancesParams   = types.NewQueryAllBalancesRequest
	ModuleCdc                   = types.ModuleCdc
	ParamStoreKeySendEnabled    = types.ParamStoreKeySendEnabled
	BalancesPrefix              = types.BalancesPrefix
	AddressFromBalancesStore    = types.AddressFromBalancesStore
)

type (
	Keeper                  = keeper.Keeper
	BaseKeeper              = keeper.BaseKeeper
	SendKeeper              = keeper.SendKeeper
	BaseSendKeeper          = keeper.BaseSendKeeper
	ViewKeeper              = keeper.ViewKeeper
	BaseViewKeeper          = keeper.BaseViewKeeper
	GenesisState            = types.GenesisState
	Balance                 = types.Balance
	MsgSend                 = types.MsgSend
	MsgMultiSend            = types.MsgMultiSend
	Input                   = types.Input
	Output                  = types.Output
	QueryBalanceRequest     = types.QueryBalanceRequest
	QueryAllBalancesRequest = types.QueryAllBalancesRequest
	GenesisBalancesIterator = types.GenesisBalancesIterator
)
