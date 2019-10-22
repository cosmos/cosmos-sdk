package keeper

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

// Keeper defines the IBC commitment keeper (i.e the vector commitment manager).
// A vector commitment manager has the ability to add or remove items from the
// commitment state as defined in https://github.com/cosmos/ics/tree/master/spec/ics-023-vector-commitments#definitions.
type Keeper struct {
	prefix              exported.PrefixI
	verifiedMemberships map[string][]byte // lookup map for returning already verified membership proofs
	verifiedAbsences    map[string]bool   // lookup map for returning already verified absences
}

// NewKeeper returns a new Keeper
func NewKeeper(prefix exported.PrefixI) Keeper {
	return Keeper{
		prefix:              prefix,
		verifiedMemberships: make(map[string][]byte),
		verifiedAbsences:    make(map[string]bool),
	}
}

// CalculateRoot returns the application Hash at the curretn block height as a commitment
// root for proof verification.
func (k Keeper) CalculateRoot(ctx sdk.Context) exported.RootI {
	return types.NewRoot(ctx.BlockHeader().AppHash)
}

// BatchVerifyMembership verifies a proof that many paths have been set to
// specific values in a commitment. It calls the proof's VerifyMembership method
// with the calculated root and the provided paths.
// Returns false on the first failed membership verification.
func (k Keeper) BatchVerifyMembership(ctx sdk.Context, proof exported.ProofI, items map[string][]byte) bool {
	root := k.CalculateRoot(ctx)

	for pathStr, value := range items {
		storedValue, ok := k.verifiedMemberships[pathStr]
		if ok && bytes.Equal(storedValue, value) {
			continue
		}

		path := types.ApplyPrefix(k.prefix, pathStr)
		ok = proof.VerifyMembership(root, path, value)
		if !ok {
			return false
		}

		k.verifiedMemberships[pathStr] = value
	}

	return true
}

// BatchVerifyNonMembership verifies a proof that many paths have not been set
// to any value in a commitment. It calls the proof's VerifyNonMembership method
// with the calculated root and the provided paths.
// Returns false on the first failed non-membership verification.
func (k Keeper) BatchVerifyNonMembership(ctx sdk.Context, proof exported.ProofI, paths []string) bool {
	root := k.CalculateRoot(ctx)
	for _, pathStr := range paths {
		ok := k.verifiedAbsences[pathStr]
		if ok {
			continue
		}

		path := types.ApplyPrefix(k.prefix, pathStr)
		ok = proof.VerifyNonMembership(root, path)
		if !ok {
			return false
		}

		k.verifiedAbsences[pathStr] = true
	}

	return true
}
