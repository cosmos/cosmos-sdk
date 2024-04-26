package expected

import (
	"context"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

type ConsensusKeeper interface {
	GetParams(ctx context.Context) (cmtproto.ConsensusParams, error)
}
