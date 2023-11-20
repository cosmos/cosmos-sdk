package keeper

import (
	"context"

	"cosmossdk.io/x/protocolpool/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) {
	for _, cf := range data.ContinuousFund {
		recipientAddress, err := k.authKeeper.AddressCodec().StringToBytes(cf.Recipient)
		if err != nil {
			panic(err)
		}
		err = k.ContinuousFund.Set(ctx, recipientAddress, cf)
		if err != nil {
			panic(err)
		}
	}
}

func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	var cf []types.ContinuousFund
	err := k.ContinuousFund.Walk(ctx, nil, func(key sdk.AccAddress, value types.ContinuousFund) (stop bool, err error) {
		cf = append(cf, types.ContinuousFund{
			Title:       value.Title,
			Description: value.Description,
			Recipient:   key.String(),
			Metadata:    value.Metadata,
			Percentage:  value.Percentage,
			Cap:         value.Cap,
			Expiry:      value.Expiry,
		})
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(cf)
}
