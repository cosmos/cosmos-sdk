package feegrant_test

import (
	"context"

	coregas "cosmossdk.io/core/gas"
	coreheader "cosmossdk.io/core/header"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type mockHeaderService struct{}

func (h mockHeaderService) HeaderInfo(ctx context.Context) coreheader.Info {
	return sdk.UnwrapSDKContext(ctx).HeaderInfo()
}

type mockGasService struct {
	coregas.Service
}

func (m mockGasService) GasMeter(ctx context.Context) coregas.Meter {
	return mockGasMeter{}
}

type mockGasMeter struct {
	coregas.Meter
}

func (m mockGasMeter) Consume(amount coregas.Gas, descriptor string) error {
	return nil
}
