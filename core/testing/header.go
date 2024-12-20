package coretesting

import (
	"context"

	"cosmossdk.io/core/header"
)

var _ header.Service = &MemHeaderService{}

type MemHeaderService struct{}

func (e MemHeaderService) HeaderInfo(ctx context.Context) header.Info {
	return unwrap(ctx).header
}
