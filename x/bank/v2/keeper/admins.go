package keeper

import (
	"context"

	"cosmossdk.io/x/bank/v2/types"
)

// GetAuthorityMetadata returns the authority metadata for a specific denom
func (k Keeper) GetAuthorityMetadata(ctx context.Context, denom string) (types.DenomAuthorityMetadata, error) {
	authority, err := k.denomAuthority.Get(ctx, denom)
	return authority, err
}

// setAuthorityMetadata stores authority metadata for a specific denom
func (k Keeper) setAuthorityMetadata(ctx context.Context, denom string, metadata types.DenomAuthorityMetadata) error {
	err := metadata.Validate()
	if err != nil {
		return err
	}

	return k.denomAuthority.Set(ctx, denom, metadata)
}

func (k Keeper) setAdmin(ctx context.Context, denom, admin string) error {
	metadata, err := k.GetAuthorityMetadata(ctx, denom)
	if err != nil {
		return err
	}

	adminAddr, err := k.addressCodec.StringToBytes(admin)
	if err != nil {
		return err
	}

	metadata.Admin = adminAddr

	return k.setAuthorityMetadata(ctx, denom, metadata)
}
