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
	ci, ok := ctx.Value(corecontext.CometInfoKey).(comet.Info)
	if !ok {
		panic("comet.Info not found in context")
	}
	return ci
}
