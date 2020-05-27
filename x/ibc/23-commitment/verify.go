package commitment

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

// CalculateRoot returns the application Hash at the curretn block height as a commitment
// root for proof verification.
func CalculateRoot(ctx sdk.Context) exported.Root {
	root := types.NewMerkleRoot(ctx.BlockHeader().AppHash)
	return &root
}

// BatchVerifyMembership verifies a proof that many paths have been set to
// specific values in a commitment. It calls the proof's VerifyMembership method
// with the calculated root and the provided paths.
// Returns false on the first failed membership verification.
func BatchVerifyMembership(
	ctx sdk.Context,
	proof exported.Proof,
	prefix exported.Prefix,
	items map[string][]byte,
) error {
	root := CalculateRoot(ctx)

	for pathStr, value := range items {
		path, err := types.ApplyPrefix(prefix, pathStr)
		if err != nil {
			return err
		}

		if err := proof.VerifyMembership(root, path, value); err != nil {
			return err
		}
	}

	return nil
}

// BatchVerifyNonMembership verifies a proof that many paths have not been set
// to any value in a commitment. It calls the proof's VerifyNonMembership method
// with the calculated root and the provided paths.
// Returns false on the first failed non-membership verification.
func BatchVerifyNonMembership(
	ctx sdk.Context,
	proof exported.Proof,
	prefix exported.Prefix,
	paths []string,
) error {
	root := CalculateRoot(ctx)
	for _, pathStr := range paths {
		path, err := types.ApplyPrefix(prefix, pathStr)
		if err != nil {
			return err
		}

		if err := proof.VerifyNonMembership(root, path); err != nil {
			return err
		}
	}

	return nil
}
