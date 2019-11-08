package evidence

import (
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"
)

// nolint

const (
	ModuleName                  = types.ModuleName
	StoreKey                    = types.StoreKey
	RouterKey                   = types.RouterKey
	QuerierRoute                = types.QuerierRoute
	DefaultParamspace           = types.DefaultParamspace
	QueryEvidence               = types.QueryEvidence
	QueryAllEvidence            = types.QueryAllEvidence
	CodeNoEvidenceHandlerExists = types.CodeNoEvidenceHandlerExists
	CodeInvalidEvidence         = types.CodeInvalidEvidence
	CodeNoEvidenceExists        = types.CodeNoEvidenceExists
	TypeMsgSubmitEvidence       = types.TypeMsgSubmitEvidence
	DefaultCodespace            = types.DefaultCodespace
	EventTypeSubmitEvidence     = types.EventTypeSubmitEvidence
	AttributeValueCategory      = types.AttributeValueCategory
	AttributeKeyEvidenceHash    = types.AttributeKeyEvidenceHash
)

var (
	NewKeeper  = keeper.NewKeeper
	NewQuerier = keeper.NewQuerier

	NewMsgSubmitEvidence      = types.NewMsgSubmitEvidence
	NewRouter                 = types.NewRouter
	NewQueryEvidenceParams    = types.NewQueryEvidenceParams
	NewQueryAllEvidenceParams = types.NewQueryAllEvidenceParams
	RegisterCodec             = types.RegisterCodec
	RegisterEvidenceTypeCodec = types.RegisterEvidenceTypeCodec
	ModuleCdc                 = types.ModuleCdc
	NewGenesisState           = types.NewGenesisState
	DefaultGenesisState       = types.DefaultGenesisState
)

type (
	Keeper = keeper.Keeper

	GenesisState      = types.GenesisState
	MsgSubmitEvidence = types.MsgSubmitEvidence
	Handler           = types.Handler
	Router            = types.Router
)
