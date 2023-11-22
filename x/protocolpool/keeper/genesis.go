package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/x/protocolpool/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) error {
	for _, cf := range data.ContinuousFund {
		recipientAddress, err := k.authKeeper.AddressCodec().StringToBytes(cf.Recipient)
		if err != nil {
			return fmt.Errorf("failed to decode recipient address: %w", err)
		}
		err = k.ContinuousFund.Set(ctx, recipientAddress, cf)
		if err != nil {
			return fmt.Errorf("failed to set continuous fund: %w", err)
		}
	}
	return nil
}

func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	var cf []types.ContinuousFund
	err := k.ContinuousFund.Walk(ctx, nil, func(key sdk.AccAddress, value types.ContinuousFund) (stop bool, err error) {
		cf = append(cf, types.ContinuousFund{
			Recipient:  key.String(),
			Percentage: value.Percentage,
			Cap:        value.Cap,
			Expiry:     value.Expiry,
		})
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(cf)
}
