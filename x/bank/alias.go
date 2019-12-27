package bank

// nolint

import (
	"github.com/cosmos/cosmos-sdk/x/bank/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/types"
)

const (
	QueryBalance       = keeper.QueryBalance
	ModuleName         = types.ModuleName
	QuerierRoute       = types.QuerierRoute
	RouterKey          = types.RouterKey
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
	NewQuerier                  = keeper.NewQuerier
	RegisterCodec               = types.RegisterCodec
	ErrNoInputs                 = types.ErrNoInputs
	ErrNoOutputs                = types.ErrNoOutputs
	ErrInputOutputMismatch      = types.ErrInputOutputMismatch
	ErrSendDisabled             = types.ErrSendDisabled
	NewGenesisState             = types.NewGenesisState
	DefaultGenesisState         = types.DefaultGenesisState
	ValidateGenesis             = types.ValidateGenesis
	NewMsgSend                  = types.NewMsgSend
	NewMsgMultiSend             = types.NewMsgMultiSend
	NewInput                    = types.NewInput
	NewOutput                   = types.NewOutput
	ValidateInputsOutputs       = types.ValidateInputsOutputs
	ParamKeyTable               = types.ParamKeyTable
	NewQueryBalanceParams       = types.NewQueryBalanceParams
	ModuleCdc                   = types.ModuleCdc
	ParamStoreKeySendEnabled    = types.ParamStoreKeySendEnabled
)

type (
	Keeper             = keeper.Keeper
	BaseKeeper         = keeper.BaseKeeper
	SendKeeper         = keeper.SendKeeper
	BaseSendKeeper     = keeper.BaseSendKeeper
	ViewKeeper         = keeper.ViewKeeper
	BaseViewKeeper     = keeper.BaseViewKeeper
	GenesisState       = types.GenesisState
	MsgSend            = types.MsgSend
	MsgMultiSend       = types.MsgMultiSend
	Input              = types.Input
	Output             = types.Output
	QueryBalanceParams = types.QueryBalanceParams
)
