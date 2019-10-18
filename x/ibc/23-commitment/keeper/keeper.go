package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

// Keeper defines the IBC commitment keeper (i.e the vector commitment manager).
// A vector commitment manager has the ability to add or remove items from the
// commitment state as defined in https://github.com/cosmos/ics/tree/master/spec/ics-023-vector-commitments#definitions.
type Keeper struct {
	prefix   exported.PrefixI
	proofs   map[string]exported.ProofI
	verified map[string][]byte
}

// NewKeeper returns a prefixed store given base store and prefix.
func NewKeeper(prefix exported.PrefixI, proofs []exported.ProofI) Keeper {
	return Keeper{
		prefix:   prefix,
		proofs:   make(map[string]exported.ProofI),
		verified: make(map[string][]byte),
	}
}

// GetRoot returns the application Hash at the curretn block height as a commitment
// root for proof verification.
func (k Keeper) GetRoot(ctx sdk.Context) exported.RootI {
	return types.NewRoot(ctx.BlockHeader().AppHash)
}

// // NewStore constructs a new Store with the root, path, and proofs.
// // The result store will be stored in the context and used by the
// // commitment.Value types.
// func NewStore(root RootI, prefix PrefixI, proofs []ProofI) (StoreI, error) {
// 	if root.CommitmentType() != prefix.CommitmentType() {
// 		return nil, errors.New("prefix type not matching with root's")
// 	}

// 	res := &store{
// 		root:     root,
// 		prefix:   prefix,
// 		proofs:   make(map[string]ProofI),
// 		verified: make(map[string][]byte),
// 	}

// 	for _, proof := range proofs {
// 		if proof.CommitmentType() != root.CommitmentType() {
// 			return nil, errors.New("proof type not matching with root's")
// 		}
// 		res.proofs[string(proof.GetKey())] = proof
// 	}

// 	return res, nil
// }

// // Prove implements spec:verifyMembership and spec:verifyNonMembership.
// // The path should be one of the path format defined under
// // https://github.com/cosmos/ics/tree/master/spec/ics-024-host-requirements
// // Prove retrieves the matching proof with the provided path from the internal map
// // and call Verify method on it with internal Root and Prefix.
// // Prove acts as verifyMembership if value is not nil, and verifyNonMembership if nil.
// func (store *store) Prove(path, value []byte) bool {
// 	stored, ok := store.verified[string(path)]
// 	if ok && bytes.Equal(stored, value) {
// 		return true
// 	}
// 	proof, ok := store.proofs[string(path)]
// 	if !ok {
// 		return false
// 	}

// 	err := proof.Verify(store.root, store.prefix, value)
// 	if err != nil {
// 		return false
// 	}
// 	store.verified[string(path)] = value

// 	return true
// }
