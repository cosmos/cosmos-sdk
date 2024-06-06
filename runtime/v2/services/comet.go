package services

import (
	"context"

	"cosmossdk.io/core/comet"
	corecontext "cosmossdk.io/core/context"
)

var _ comet.Service = &ContextAwareCometInfoService{}

type ContextAwareCometInfoService struct{}

// CometInfo implements comet.Service.
func (c *ContextAwareCometInfoService) CometInfo(ctx context.Context) comet.Info {
	v := ctx.Value(corecontext.CometInfoKey)
	ci := v.(comet.Info)
	return ci
}
