package exported

import (
	"context"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

type (

	// ConsensusParamSetter defines the interface fulfilled by BaseApp's
	// ParamStore which allows setting its appVersion field.
	ConsensusParamSetter interface {
		Get(ctx context.Context) (cmtproto.ConsensusParams, error)
		Has(ctx context.Context) (bool, error)
		Set(ctx context.Context, cp cmtproto.ConsensusParams) error
	}
)
