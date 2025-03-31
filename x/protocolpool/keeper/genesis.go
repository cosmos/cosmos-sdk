package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) error {
	currentTime := ctx.BlockTime()

	err := k.Params.Set(ctx, data.Params)
	if err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	for _, cf := range data.ContinuousFunds {
		recipientAddress, err := k.authKeeper.AddressCodec().StringToBytes(cf.Recipient)
		if err != nil {
			return fmt.Errorf("failed to decode recipient address: %w", err)
		}

		if k.bankKeeper.BlockedAddr(recipientAddress) {
			return fmt.Errorf("recipient is blocked in the bank keeper: %s", recipientAddress)
		}

		// ignore expired ContinuousFunds
		if cf.Expiry != nil && cf.Expiry.Before(currentTime) {
			continue
		}

		if err := k.ContinuousFunds.Set(ctx, recipientAddress, cf); err != nil {
			return fmt.Errorf("failed to set continuous fund for recipient %s: %w", recipientAddress, err)
		}
	}

	return nil
}

func (k Keeper) ExportGenesis(ctx sdk.Context) (*types.GenesisState, error) {
	cf, err := k.GetAllContinuousFunds(ctx)
	if err != nil {
		return nil, err
	}

	genState := types.NewGenesisState(cf)

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	genState.Params = params

	return genState, nil
}
