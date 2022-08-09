package baseapp

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/store/v2alpha1/smt"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
)

// Represents a single-round fraudProof
type FraudProof struct {
	// The block height to load state of
	blockHeight int64

	// A map from module name to state witness
	stateWitness map[string]StateWitness
}

// State witness with a list of all witness data
type StateWitness struct {
	// store level proof
	proof    tmcrypto.ProofOp
	rootHash []byte
	// List of witness data
	WitnessData []WitnessData
}

// Witness data containing a key/value pair and a SMT proof for said key/value pair
type WitnessData struct {
	Key   []byte
	Value []byte
	proof tmcrypto.ProofOp
}

func (fraudProof *FraudProof) verifyFraudProof(headerAppHash []byte) (bool, error) {
	for storeKey, stateWitness := range fraudProof.stateWitness {
		proofOp := stateWitness.proof
		proof, err := types.CommitmentOpDecoder(proofOp)
		if err != nil {
			return false, err
		}
		if !bytes.Equal(proof.GetKey(), []byte(storeKey)) {
			return false, fmt.Errorf("got storeKey: %s, expected: %s", string(proof.GetKey()), storeKey)
		}
		appHash, err := proof.Run([][]byte{stateWitness.rootHash})
		if err != nil {
			return false, err
		}
		if !bytes.Equal(appHash[0], headerAppHash) {
			return false, fmt.Errorf("got appHash: %s, expected: %s", string(headerAppHash), string(headerAppHash))
		}

		// Fraudproof verification on a substore level
		for _, witness := range stateWitness.WitnessData {
			proofOp, key, value := witness.proof, witness.Key, witness.Value
			proof, err := smt.ProofDecoder(proofOp)
			if err != nil {
				return false, err
			}
			if !bytes.Equal(key, proof.GetKey()) {
				return false, fmt.Errorf("got key: %s, expected: %s for storeKey: %s", string(key), string(proof.GetKey()), storeKey)
			}
			rootHash, err := proof.Run([][]byte{value})
			if err != nil {
				return false, err
			}
			if !bytes.Equal(rootHash[0], stateWitness.rootHash) {
				return false, fmt.Errorf("got rootHash: %s, expected: %s for storeKey: %s", string(rootHash[0]), string(stateWitness.rootHash), storeKey)
			}
		}
	}
	return true, nil
}
