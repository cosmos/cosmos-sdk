package evidence

import (
	"github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

// nolint

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

	NewMsgSubmitEvidenceBase     = types.NewMsgSubmitEvidenceBase
	NewRouter                    = types.NewRouter
	NewQueryEvidenceParams       = types.NewQueryEvidenceParams
	NewQueryAllEvidenceParams    = types.NewQueryAllEvidenceParams
	RegisterCodec                = types.RegisterCodec
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

	GenesisState          = types.GenesisState
	MsgSubmitEvidenceBase = types.MsgSubmitEvidenceBase
	Handler               = types.Handler
	Router                = types.Router
	Equivocation          = types.Equivocation
	Codec                 = types.Codec
)
