package stf

import (
	"context"

	"cosmossdk.io/core/header"
)

var _ header.Service = (*HeaderService)(nil)

type HeaderService struct {
	getHeader func() (header.Info, error)
}

func (h HeaderService) HeaderInfo(ctx context.Context) header.Info {
	return ctx.(*executionContext).headerInfo
}
