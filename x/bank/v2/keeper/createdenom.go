package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/x/bank/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// ConvertToBaseToken converts a fee amount in a whitelisted fee token to the base fee token amount
func (k Keeper) CreateDenom(ctx context.Context, creatorAddr string, subdenom string) (newTokenDenom string, err error) {
	denom, err := k.validateCreateDenom(ctx, creatorAddr, subdenom)
	if err != nil {
		return "", err
	}

	err = k.chargeForCreateDenom(ctx, creatorAddr)
	if err != nil {
		return "", err
	}

	err = k.createDenomAfterValidation(ctx, creatorAddr, denom)
	return denom, err
}

// Runs CreateDenom logic after the charge and all denom validation has been handled.
// Made into a second function for genesis initialization.
func (k Keeper) createDenomAfterValidation(ctx context.Context, creatorAddr string, denom string) (err error) {
	_, exists := k.GetDenomMetaData(ctx, denom)
	if !exists {
		denomMetaData := types.Metadata{
			DenomUnits: []*types.DenomUnit{{
				Denom:    denom,
				Exponent: 0,
			}},
			Base:    denom,
			Name:    denom,
			Symbol:  denom,
			Display: denom,
		}

		err := k.SetDenomMetaData(ctx, denomMetaData)
		if err != nil {
			return err
		}
	}

	authorityMetadata := types.DenomAuthorityMetadata{
		Admin: creatorAddr,
	}
	err = k.setAuthorityMetadata(ctx, denom, authorityMetadata)
	if err != nil {
		return err
	}

	// TODO: do we need map creator => denom
	// k.addDenomFromCreator(ctx, creatorAddr, denom)
	return nil
}

func (k Keeper) validateCreateDenom(ctx context.Context, creatorAddr string, subdenom string) (newTokenDenom string, err error) {
	// Temporary check until IBC bug is sorted out
	if k.HasSupply(ctx, subdenom) {
		return "", fmt.Errorf("temporary error until IBC bug is sorted out, " +
			"can't create subdenoms that are the same as a native denom")
	}

	denom, err := types.GetTokenDenom(creatorAddr, subdenom)
	if err != nil {
		return "", err
	}

	_, found := k.GetDenomMetaData(ctx, denom)
	if found {
		return "", types.ErrDenomExists
	}

	return denom, nil
}

func (k Keeper) chargeForCreateDenom(ctx context.Context, creatorAddr string) (err error) {
	params := k.GetParams(ctx)

	// if DenomCreationFee is non-zero, transfer the tokens from the creator
	// account to community pool
	if params.DenomCreationFee != nil {
		accAddr, err := sdk.AccAddressFromBech32(creatorAddr)
		if err != nil {
			return err
		}

		communityPoolAddr := address.Module("protocolpool")

		if err := k.SendCoins(ctx, accAddr, communityPoolAddr, params.DenomCreationFee); err != nil {
			return err
		}
	}

	// if DenomCreationGasConsume is non-zero, consume the gas
	if params.DenomCreationGasConsume != 0 {
		err = k.Environment.GasService.GasMeter(ctx).Consume(params.DenomCreationGasConsume, "consume denom creation gas")
		if err != nil {
			return err
		}
	}

	return nil
}
