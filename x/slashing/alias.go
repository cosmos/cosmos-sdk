package slashing

// nolint

import (
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/types"
)

const (
	ModuleName                  = types.ModuleName
	StoreKey                    = types.StoreKey
	RouterKey                   = types.RouterKey
	QuerierRoute                = types.QuerierRoute
	DefaultParamspace           = types.DefaultParamspace
	DefaultSignedBlocksWindow   = types.DefaultSignedBlocksWindow
	DefaultDowntimeJailDuration = types.DefaultDowntimeJailDuration
	QueryParameters             = types.QueryParameters
	QuerySigningInfo            = types.QuerySigningInfo
	QuerySigningInfos           = types.QuerySigningInfos

	EventTypeSlash                 = types.EventTypeSlash
	EventTypeLiveness              = types.EventTypeLiveness
	AttributeKeyAddress            = types.AttributeKeyAddress
	AttributeKeyHeight             = types.AttributeKeyHeight
	AttributeKeyPower              = types.AttributeKeyPower
	AttributeKeyReason             = types.AttributeKeyReason
	AttributeKeyJailed             = types.AttributeKeyJailed
	AttributeKeyMissedBlocks       = types.AttributeKeyMissedBlocks
	AttributeValueDoubleSign       = types.AttributeValueDoubleSign
	AttributeValueMissingSignature = types.AttributeValueMissingSignature
	AttributeValueCategory         = types.AttributeValueCategory
)

var (
	// functions aliases
	NewKeeper                                = keeper.NewKeeper
	NewQuerier                               = keeper.NewQuerier
	RegisterCodec                            = types.RegisterCodec
	ErrNoValidatorForAddress                 = types.ErrNoValidatorForAddress
	ErrBadValidatorAddr                      = types.ErrBadValidatorAddr
	ErrValidatorJailed                       = types.ErrValidatorJailed
	ErrValidatorNotJailed                    = types.ErrValidatorNotJailed
	ErrMissingSelfDelegation                 = types.ErrMissingSelfDelegation
	ErrSelfDelegationTooLowToUnjail          = types.ErrSelfDelegationTooLowToUnjail
	ErrNoSigningInfoFound                    = types.ErrNoSigningInfoFound
	NewGenesisState                          = types.NewGenesisState
	NewMissedBlock                           = types.NewMissedBlock
	DefaultGenesisState                      = types.DefaultGenesisState
	ValidateGenesis                          = types.ValidateGenesis
	GetValidatorSigningInfoKey               = types.GetValidatorSigningInfoKey
	GetValidatorSigningInfoAddress           = types.GetValidatorSigningInfoAddress
	GetValidatorMissedBlockBitArrayPrefixKey = types.GetValidatorMissedBlockBitArrayPrefixKey
	GetValidatorMissedBlockBitArrayKey       = types.GetValidatorMissedBlockBitArrayKey
	GetAddrPubkeyRelationKey                 = types.GetAddrPubkeyRelationKey
	NewMsgUnjail                             = types.NewMsgUnjail
	ParamKeyTable                            = types.ParamKeyTable
	NewParams                                = types.NewParams
	DefaultParams                            = types.DefaultParams
	NewQuerySigningInfoParams                = types.NewQuerySigningInfoParams
	NewQuerySigningInfosParams               = types.NewQuerySigningInfosParams
	NewValidatorSigningInfo                  = types.NewValidatorSigningInfo

	// variable aliases
	ModuleCdc                       = types.ModuleCdc
	ValidatorSigningInfoKey         = types.ValidatorSigningInfoKey
	ValidatorMissedBlockBitArrayKey = types.ValidatorMissedBlockBitArrayKey
	AddrPubkeyRelationKey           = types.AddrPubkeyRelationKey
	DefaultMinSignedPerWindow       = types.DefaultMinSignedPerWindow
	DefaultSlashFractionDoubleSign  = types.DefaultSlashFractionDoubleSign
	DefaultSlashFractionDowntime    = types.DefaultSlashFractionDowntime
	KeySignedBlocksWindow           = types.KeySignedBlocksWindow
	KeyMinSignedPerWindow           = types.KeyMinSignedPerWindow
	KeyDowntimeJailDuration         = types.KeyDowntimeJailDuration
	KeySlashFractionDoubleSign      = types.KeySlashFractionDoubleSign
	KeySlashFractionDowntime        = types.KeySlashFractionDowntime
)

type (
	Hooks                   = keeper.Hooks
	Keeper                  = keeper.Keeper
	GenesisState            = types.GenesisState
	MissedBlock             = types.MissedBlock
	MsgUnjail               = types.MsgUnjail
	Params                  = types.Params
	QuerySigningInfoParams  = types.QuerySigningInfoParams
	QuerySigningInfosParams = types.QuerySigningInfosParams
	ValidatorSigningInfo    = types.ValidatorSigningInfo
)
