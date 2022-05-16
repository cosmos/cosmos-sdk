package v047

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// MigrateStore performs in-place store migrations from bank module v3 to v4.
// migration includes:
//
// - Moving the SendEnabled information from Params into the bank store.
func MigrateStore(ctx sdk.Context, keeper bankkeeper.BaseKeeper) error {
	return moveSendEnabledToStore(ctx, keeper)
}

// moveSendEnabledToStore gets the params for the bank module, sets all the SendEnabled entries in the bank store,
// then deletes all the SendEnabled entries from the params.
func moveSendEnabledToStore(ctx sdk.Context, keeper bankkeeper.BaseKeeper) error {
	params := keeper.GetParams(ctx)
	keeper.SetAllSendEnabled(ctx, params.GetSendEnabled())
	params.SendEnabled = []*banktypes.SendEnabled{}
	keeper.SetParams(ctx, params)
	return nil
}
