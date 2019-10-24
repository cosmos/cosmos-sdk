package evidence

import (
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"
)

const (
	ModuleName        = types.ModuleName
	StoreKey          = types.StoreKey
	RouterKey         = types.RouterKey
	QuerierRoute      = types.QuerierRoute
	DefaultParamspace = types.DefaultParamspace
)

var (
	RegisterCodec             = types.RegisterCodec
	RegisterEvidenceTypeCodec = types.RegisterEvidenceTypeCodec
	ModuleCdc                 = types.ModuleCdc
)

type (
	Evidence = types.Evidence
	Handler  = types.Handler
	Router   = types.Router
)
