package baseapp

import (
	"context"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

const InitialAppVersion uint64 = 0

// ParamStore defines the interface the parameter store used by the BaseApp must
// fulfill.
type ParamStore interface {
	Get(ctx context.Context) (cmtproto.ConsensusParams, error)
	Has(ctx context.Context) (bool, error)
	Set(ctx context.Context, cp cmtproto.ConsensusParams) error
}

// AppVersionModifier defines the interface fulfilled by BaseApp
// which allows getting and setting it's appVersion field. This
// in turn updates the consensus params that are sent to the
// consensus engine in EndBlock
type AppVersionModifier interface {
	SetAppVersion(context.Context, uint64) error
	AppVersion(context.Context) (uint64, error)
}
