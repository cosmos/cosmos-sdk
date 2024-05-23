package mint

import (
	"context"

	epochstypes "cosmossdk.io/x/epochs/types"
)

var _ epochstypes.EpochHooks = AppModule{}

// GetModuleName implements types.EpochHooks.
func (am AppModule) GetModuleName() string {
	return am.Name()
}

// BeforeEpochStart is a noop
func (am AppModule) BeforeEpochStart(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	minter, err := am.keeper.Minter.Get(ctx)
	if err != nil {
		return err
	}

	err = am.mintFn(ctx, am.keeper.Environment, &minter, epochIdentifier, epochNumber)
	if err != nil {
		return err
	}

	return am.keeper.Minter.Set(ctx, minter)
}

func (am AppModule) AfterEpochEnd(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}
