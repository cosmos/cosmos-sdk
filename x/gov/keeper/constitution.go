package keeper

import (
	"errors"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// ApplyConstitutionAmendment applies the amendment as a patch against the current constitution
// and returns the updated constitution. If the amendment cannot be applied cleanly, an error is returned.
func (keeper Keeper) ApplyConstitutionAmendment(ctx sdk.Context, amendment string) (updatedConstitution string, err error) {
	if amendment == "" {
		return "", types.ErrInvalidConstitutionAmendment.Wrap("amendment cannot be empty")
	}

	currentConstitution, err := keeper.Constitution.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return "", err
	}

	updatedConstitution, err = types.ApplyUnifiedDiff(currentConstitution, amendment)
	if err != nil {
		return "", types.ErrInvalidConstitutionAmendment.Wrapf("failed to apply amendment: %v", err)
	}

	return updatedConstitution, nil
}
