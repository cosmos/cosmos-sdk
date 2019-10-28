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
)

var (
	NewKeeper  = keeper.NewKeeper
	NewQuerier = keeper.NewQuerier

	NewQueryEvidenceParams    = types.NewQueryEvidenceParams
	NewQueryAllEvidenceParams = types.NewQueryAllEvidenceParams
	RegisterCodec             = types.RegisterCodec
	RegisterEvidenceTypeCodec = types.RegisterEvidenceTypeCodec
	ModuleCdc                 = types.ModuleCdc
)

type (
	Keeper = keeper.Keeper

	MsgSubmitEvidence = types.MsgSubmitEvidence
	Evidence          = types.Evidence
	Handler           = types.Handler
	Router            = types.Router
)
