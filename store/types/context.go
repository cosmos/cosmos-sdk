package types

import (
	"context"

	"github.com/tendermint/tendermint/libs/log"
)

type Context interface {
	Context() context.Context
	BlockHeight() int64
	Logger() log.Logger
	StreamingManager() StreamingManager
}
