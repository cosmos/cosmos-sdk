package evidence

import (
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"
)

// nolint

const (
	ModuleName        = types.ModuleName
	StoreKey          = types.StoreKey
	RouterKey         = types.RouterKey
	QuerierRoute      = types.QuerierRoute
	DefaultParamspace = types.DefaultParamspace
)

var (
	NewKeeper = keeper.NewKeeper

	RegisterCodec             = types.RegisterCodec
	RegisterEvidenceTypeCodec = types.RegisterEvidenceTypeCodec
	ModuleCdc                 = types.ModuleCdc
)

type (
	Keeper = keeper.Keeper

	Evidence = types.Evidence
	Handler  = types.Handler
	Router   = types.Router
)
