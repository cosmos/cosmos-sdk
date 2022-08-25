package baseapp

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/mem"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/store/v2alpha1/smt"
	smtlib "github.com/lazyledger/smt"
	abci "github.com/tendermint/tendermint/abci/types"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
)

// Represents a single-round fraudProof
type FraudProof struct {
	// The block height to load state of
	blockHeight int64

	appHash []byte
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

func (fraudProof *FraudProof) getModules() []string {
	keys := make([]string, 0, len(fraudProof.stateWitness))
	for k := range fraudProof.stateWitness {
		keys = append(keys, k)
	}
	return keys
}

func (fraudProof *FraudProof) getSubstoreSMTs() (map[string]*smtlib.SparseMerkleTree, error) {
	storeKeyToSMT := make(map[string]*smtlib.SparseMerkleTree)
	for storeKey, stateWitness := range fraudProof.stateWitness {
		rootHash := stateWitness.rootHash
		substoreDeepSMT := smtlib.NewDeepSparseMerkleSubTree(smtlib.NewSimpleMap(), smtlib.NewSimpleMap(), sha256.New(), rootHash)
		for _, witnessData := range stateWitness.WitnessData {
			proofOp, key, val := witnessData.proof, witnessData.Key, witnessData.Value
			proof, err := smt.ProofDecoder(proofOp)
			if err != nil {
				return nil, err
			}
			smtProof := proof.(*smt.ProofOp).GetProof()
			substoreDeepSMT.AddBranch(smtProof, key, val)
		}
		storeKeyToSMT[storeKey] = substoreDeepSMT.SparseMerkleTree
	}
	return storeKeyToSMT, nil
}

func (fraudProof *FraudProof) extractStore() map[string]types.KVStore {
	store := make(map[string]types.KVStore)
	for storeKey, stateWitness := range fraudProof.stateWitness {
		subStore := mem.NewStore()
		for _, witnessData := range stateWitness.WitnessData {
			key, val := witnessData.Key, witnessData.Value
			subStore.Set(key, val)
		}
		store[storeKey] = subStore
	}
	return store
}

func (fraudProof *FraudProof) verifyFraudProof() (bool, error) {
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
		if !bytes.Equal(appHash[0], fraudProof.appHash) {
			return false, fmt.Errorf("got appHash: %s, expected: %s", string(fraudProof.appHash), string(fraudProof.appHash))
		}
		// Fraudproof verification on a substore level
		for _, witness := range stateWitness.WitnessData {
			proofOp, key, value := witness.proof, witness.Key, witness.Value
			if err != nil {
				return false, err
			}
			if !bytes.Equal(key, proofOp.GetKey()) {
				return false, fmt.Errorf("got key: %s, expected: %s for storeKey: %s", string(key), string(proof.GetKey()), storeKey)
			}
			proof, err := smt.ProofDecoder(proofOp)
			if err != nil {
				return false, err
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

func (fraudProof *FraudProof) toABCI() abci.FraudProof {
	abciStateWitness := make(map[string]*abci.StateWitness)
	for storeKey, stateWitness := range fraudProof.stateWitness {
		abciWitnessData := make([]*abci.WitnessData, len(stateWitness.WitnessData))
		for _, witnessData := range stateWitness.WitnessData {
			abciWitness := abci.WitnessData{
				Key:   witnessData.Key,
				Value: witnessData.Value,
				Proof: &witnessData.proof,
			}
			abciWitnessData = append(abciWitnessData, &abciWitness)
		}
		abciStateWitness[storeKey] = &abci.StateWitness{
			ProofOp:     &stateWitness.proof,
			RootHash:    stateWitness.rootHash,
			WitnessData: abciWitnessData,
		}
	}
	return abci.FraudProof{
		BlockHeight:  fraudProof.blockHeight,
		AppHash:      fraudProof.appHash,
		StateWitness: abciStateWitness,
	}
}
