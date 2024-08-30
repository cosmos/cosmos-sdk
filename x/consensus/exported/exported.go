package exported

import (
	"context"

	"cosmossdk.io/x/consensus/types"
)

// ConsensusParamSetter defines the interface fulfilled by BaseApp's
// ParamStore which allows setting its appVersion field.
type ConsensusParamSetter interface {
	Get(ctx context.Context) (types.ConsensusParams, error)
	Has(ctx context.Context) (bool, error)
	Set(ctx context.Context, cp types.ConsensusParams) error
}
