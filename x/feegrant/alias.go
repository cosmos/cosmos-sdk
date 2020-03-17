package feegrant

import (
	"github.com/cosmos/cosmos-sdk/x/feegrant/ante"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

// nolint

const (
	DefaultCodespace        = types.DefaultCodespace
	EventTypeUseFeeGrant    = types.EventTypeUseFeeGrant
	EventTypeRevokeFeeGrant = types.EventTypeRevokeFeeGrant
	EventTypeSetFeeGrant    = types.EventTypeSetFeeGrant
	AttributeKeyGranter     = types.AttributeKeyGranter
	AttributeKeyGrantee     = types.AttributeKeyGrantee
	ModuleName              = types.ModuleName
	StoreKey                = types.StoreKey
	RouterKey               = types.RouterKey
	QuerierRoute            = types.QuerierRoute
	QueryGetFeeAllowances   = keeper.QueryGetFeeAllowances
)

var (
	NewDeductGrantedFeeDecorator = ante.NewDeductGrantedFeeDecorator
	RegisterCodec                = types.RegisterCodec
	ExpiresAtTime                = types.ExpiresAtTime
	ExpiresAtHeight              = types.ExpiresAtHeight
	ClockDuration                = types.ClockDuration
	BlockDuration                = types.BlockDuration
	FeeAllowanceKey              = types.FeeAllowanceKey
	FeeAllowancePrefixByGrantee  = types.FeeAllowancePrefixByGrantee
	NewMsgRevokeFeeAllowance     = types.NewMsgRevokeFeeAllowance
	NewFeeGrantTx                = types.NewFeeGrantTx
	NewMsgGrantFeeAllowanceBase  = types.NewMsgGrantFeeAllowanceBase
	NewFeeAllowanceGrantBase     = types.NewFeeAllowanceGrantBase
	CountSubKeys                 = types.CountSubKeys
	NewGrantedFee                = types.NewGrantedFee
	StdSignBytes                 = types.StdSignBytes
	NewKeeper                    = keeper.NewKeeper
	NewQuerier                   = keeper.NewQuerier

	ModuleCdc             = types.ModuleCdc
	ErrFeeLimitExceeded   = types.ErrFeeLimitExceeded
	ErrFeeLimitExpired    = types.ErrFeeLimitExpired
	ErrInvalidDuration    = types.ErrInvalidDuration
	ErrNoAllowance        = types.ErrNoAllowance
	FeeAllowanceKeyPrefix = types.FeeAllowanceKeyPrefix
)

type (
	GrantedFeeTx              = ante.GrantedFeeTx
	DeductGrantedFeeDecorator = ante.DeductGrantedFeeDecorator
	BasicFeeAllowance         = types.BasicFeeAllowance
	ExpiresAt                 = types.ExpiresAt
	Duration                  = types.Duration
	FeeAllowanceGrantBase     = types.FeeAllowanceGrantBase
	MsgGrantFeeAllowanceBase  = types.MsgGrantFeeAllowanceBase
	MsgRevokeFeeAllowance     = types.MsgRevokeFeeAllowance
	PeriodicFeeAllowance      = types.PeriodicFeeAllowance
	FeeGrantTx                = types.FeeGrantTx
	GrantedFee                = types.GrantedFee
	DelegatedSignDoc          = types.DelegatedSignDoc
	Keeper                    = keeper.Keeper
	Codec                     = types.Codec
)
