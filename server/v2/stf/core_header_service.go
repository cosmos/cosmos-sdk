package stf

import (
	"context"

	"cosmossdk.io/core/header"
)

var _ header.Service = (*HeaderService)(nil)

type HeaderService struct {
	Info header.Info
}

func (h HeaderService) GetHeaderInfo(ctx context.Context) header.Info {
	return h.Info
}
