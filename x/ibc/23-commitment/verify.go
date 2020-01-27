package commitment

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CalculateRoot returns the application Hash at the curretn block height as a commitment
// root for proof verification.
func CalculateRoot(ctx sdk.Context) RootI {
	return NewRoot(ctx.BlockHeader().AppHash)
}

// BatchVerifyMembership verifies a proof that many paths have been set to
// specific values in a commitment. It calls the proof's VerifyMembership method
// with the calculated root and the provided paths.
// Returns false on the first failed membership verification.
func BatchVerifyMembership(
	ctx sdk.Context,
	proof ProofI,
	prefix PrefixI,
	items map[string][]byte,
) error {
	root := CalculateRoot(ctx)

	for pathStr, value := range items {
		path, err := ApplyPrefix(prefix, pathStr)
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
	proof ProofI,
	prefix PrefixI,
	paths []string,
) error {
	root := CalculateRoot(ctx)
	for _, pathStr := range paths {
		path, err := ApplyPrefix(prefix, pathStr)
		if err != nil {
			return err
		}

		if err := proof.VerifyNonMembership(root, path); err != nil {
			return err
		}
	}

	return nil
}
