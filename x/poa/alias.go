package poa

import (
	"github.com/cosmos/cosmos-sdk/x/poa/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/poa/internal/types"
)

const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	TStoreKey    = types.TStoreKey
	QuerierRoute = types.QuerierRoute
	RouterKey    = types.RouterKey
)

var (
	RegisterInvariants = keeper.RegisterInvariants
	NewKeeper          = keeper.NewKeeper
	ParamKeyTable      = keeper.ParamKeyTable
	NewQuerier         = keeper.NewQuerier

	NewValidator  = types.NewValidator
	RegisterCodec = types.RegisterCodec
	ModuleCdc     = types.ModuleCdc

	EventTypeCompleteUnbonding = types.EventTypeCompleteUnbonding
	EventTypeCreateValidator   = types.EventTypeCreateValidator
	EventTypeEditValidator     = types.EventTypeEditValidator
	EventTypeUnbond            = types.EventTypeUnbond
	AttributeKeyValidator      = types.AttributeKeyValidator
	AttributeKeySrcValidator   = types.AttributeKeySrcValidator
	AttributeKeyCompletionTime = types.AttributeKeyCompletionTime
	AttributeValueCategory     = types.AttributeValueCategory
	DefaultGenesisState        = types.DefaultGenesisState
	NewGenesisState            = types.NewGenesisState
)

type (
	Keeper = keeper.Keeper

	Validator                 = types.Validator
	MsgCreateValidator        = types.MsgCreateValidator
	MsgProposeCreateValidator = types.MsgProposeCreateValidator
	MsgEditValidator          = types.MsgEditValidator
	GenesisState              = types.GenesisState
	SupplyKeeper              = types.SupplyKeeper
	AccountKeeper             = types.AccountKeeper
)
