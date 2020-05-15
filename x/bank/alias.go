package bank

import (
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	QueryBalance       = types.QueryBalance
	QueryAllBalances   = types.QueryAllBalances
	DefaultParamspace  = types.DefaultParamspace
	DefaultSendEnabled = types.DefaultSendEnabled

	EventTypeTransfer      = types.EventTypeTransfer
	AttributeKeyRecipient  = types.AttributeKeyRecipient
	AttributeKeySender     = types.AttributeKeySender
	AttributeValueCategory = types.AttributeValueCategory

	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	RouterKey    = types.RouterKey
	QuerierRoute = types.QuerierRoute
)

var (
	RegisterInvariants          = keeper.RegisterInvariants
	NonnegativeBalanceInvariant = keeper.NonnegativeBalanceInvariant
	NewBaseKeeper               = keeper.NewBaseKeeper
	NewBaseSendKeeper           = keeper.NewBaseSendKeeper
	NewBaseViewKeeper           = keeper.NewBaseViewKeeper
	NewQuerier                  = keeper.NewQuerier
	RegisterCodec               = types.RegisterCodec
	ErrNoInputs                 = types.ErrNoInputs
	ErrNoOutputs                = types.ErrNoOutputs
	ErrInputOutputMismatch      = types.ErrInputOutputMismatch
	ErrSendDisabled             = types.ErrSendDisabled
	NewGenesisState             = types.NewGenesisState
	DefaultGenesisState         = types.DefaultGenesisState
	SanitizeGenesisBalances     = types.SanitizeGenesisBalances
	GetGenesisStateFromAppState = types.GetGenesisStateFromAppState
	NewMsgSend                  = types.NewMsgSend
	NewMsgMultiSend             = types.NewMsgMultiSend
	NewInput                    = types.NewInput
	NewOutput                   = types.NewOutput
	ValidateInputsOutputs       = types.ValidateInputsOutputs
	ParamKeyTable               = types.ParamKeyTable
	NewQueryBalanceParams       = types.NewQueryBalanceParams
	NewQueryAllBalancesParams   = types.NewQueryAllBalancesParams
	ModuleCdc                   = types.ModuleCdc
	ParamStoreKeySendEnabled    = types.ParamStoreKeySendEnabled
	BalancesPrefix              = types.BalancesPrefix
	AddressFromBalancesStore    = types.AddressFromBalancesStore
	AllInvariants               = keeper.AllInvariants
	TotalSupply                 = keeper.TotalSupply
	NewSupply                   = types.NewSupply
	DefaultSupply               = types.DefaultSupply
)

type (
	BaseKeeper              = keeper.BaseKeeper
	SendKeeper              = keeper.SendKeeper
	BaseSendKeeper          = keeper.BaseSendKeeper
	ViewKeeper              = keeper.ViewKeeper
	BaseViewKeeper          = keeper.BaseViewKeeper
	Balance                 = types.Balance
	MsgSend                 = types.MsgSend
	MsgMultiSend            = types.MsgMultiSend
	Input                   = types.Input
	Output                  = types.Output
	QueryBalanceParams      = types.QueryBalanceParams
	QueryAllBalancesParams  = types.QueryAllBalancesParams
	GenesisBalancesIterator = types.GenesisBalancesIterator
	Keeper                  = keeper.Keeper
	GenesisState            = types.GenesisState
	Supply                  = types.Supply
)
