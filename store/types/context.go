package types

import (
	"cosmossdk.io/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

// Context is an interface used by an App to pass context information
// needed to process store streaming requests.
type Context interface {
	BlockHeader() tmproto.Header
	BlockHeight() int64
	Logger() log.Logger
	StreamingManager() StreamingManager
}
