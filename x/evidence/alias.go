package evidence

import (
	"github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

const (
	ModuleName               = types.ModuleName
	StoreKey                 = types.StoreKey
	RouterKey                = types.RouterKey
	QuerierRoute             = types.QuerierRoute
	QueryEvidence            = types.QueryEvidence
	QueryAllEvidence         = types.QueryAllEvidence
	TypeMsgSubmitEvidence    = types.TypeMsgSubmitEvidence
	EventTypeSubmitEvidence  = types.EventTypeSubmitEvidence
	AttributeValueCategory   = types.AttributeValueCategory
	AttributeKeyEvidenceHash = types.AttributeKeyEvidenceHash
)

var (
	NewKeeper  = keeper.NewKeeper
	NewQuerier = keeper.NewQuerier

	NewMsgSubmitEvidence         = types.NewMsgSubmitEvidence
	NewRouter                    = types.NewRouter
	NewQueryEvidenceParams       = types.NewQueryEvidenceParams
	NewQueryAllEvidenceParams    = types.NewQueryAllEvidenceParams
	RegisterCodec                = types.RegisterCodec
	RegisterInterfaces           = types.RegisterInterfaces
	ModuleCdc                    = types.ModuleCdc
	NewGenesisState              = types.NewGenesisState
	DefaultGenesisState          = types.DefaultGenesisState
	ConvertDuplicateVoteEvidence = types.ConvertDuplicateVoteEvidence
	DoubleSignJailEndTime        = types.DoubleSignJailEndTime
	ErrNoEvidenceHandlerExists   = types.ErrNoEvidenceHandlerExists
	ErrInvalidEvidence           = types.ErrInvalidEvidence
	ErrNoEvidenceExists          = types.ErrNoEvidenceExists
	ErrEvidenceExists            = types.ErrEvidenceExists
)

type (
	Keeper = keeper.Keeper

	GenesisState      = types.GenesisState
	MsgSubmitEvidence = types.MsgSubmitEvidence
	Handler           = types.Handler
	Router            = types.Router
	Equivocation      = types.Equivocation
)
