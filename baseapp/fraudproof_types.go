package baseapp

import (
	"bytes"
	"errors"
	"fmt"

	ics23 "github.com/confio/ics23/go"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/iavl"
	abci "github.com/tendermint/tendermint/abci/types"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
	db "github.com/tendermint/tm-db"
)

var ErrMoreThanOneBlockTypeUsed = errors.New("fraud proof has not exactly one type of fraudulent state transitions marked nil")

// FraudProof represents a single-round fraudProof
type FraudProof struct {
	// The block height to load state of, aka the last committed block. Note: this diverges from the ADR
	BlockHeight int64

	PreStateAppHash      []byte
	ExpectedValidAppHash []byte
	// A map from module name to state witness
	stateWitness map[string]StateWitness

	// Fraudulent state transition has to be one of these
	// Only one of these three can be non-nil
	FraudulentBeginBlock *abci.RequestBeginBlock
	FraudulentDeliverTx  *abci.RequestDeliverTx
	FraudulentEndBlock   *abci.RequestEndBlock
}

// StateWitness with a list of all witness data
type StateWitness struct {
	// store level proof
	Proof    tmcrypto.ProofOp
	RootHash []byte
	// List of witness data
	WitnessData []*WitnessData
}

// WitnessData represents a trace operation along with inclusion proofs required for said operation
type WitnessData struct {
	Operation iavl.Operation
	Key       []byte
	Value     []byte
	Proofs    []*tmcrypto.ProofOp
}

func convertToProofOps(existenceProofs []*ics23.ExistenceProof) []*tmcrypto.ProofOp {
	if existenceProofs == nil {
		return nil
	}
	proofOps := make([]*tmcrypto.ProofOp, 0)
	for _, existenceProof := range existenceProofs {
		proofOps = append(proofOps, getProofOp(existenceProof))
	}
	return proofOps
}

func getProofOp(exist *ics23.ExistenceProof) *tmcrypto.ProofOp {
	commitmentProof := &ics23.CommitmentProof{
		Proof: &ics23.CommitmentProof_Exist{
			Exist: exist,
		},
	}
	proofOp := types.NewIavlCommitmentOp(exist.Key, commitmentProof).ProofOp()
	return &proofOp
}

func convertToExistenceProofs(proofs []*tmcrypto.ProofOp) ([]*ics23.ExistenceProof, error) {
	existenceProofs := make([]*ics23.ExistenceProof, 0)
	for _, proof := range proofs {
		_, existenceProof, err := getExistenceProof(*proof)
		if err != nil {
			return nil, err
		}
		existenceProofs = append(existenceProofs, existenceProof)
	}
	return existenceProofs, nil
}

func getExistenceProof(proofOp tmcrypto.ProofOp) (types.CommitmentOp, *ics23.ExistenceProof, error) {
	op, err := types.CommitmentOpDecoder(proofOp)
	if err != nil {
		return types.CommitmentOp{}, nil, err
	}
	commitmentOp := op.(types.CommitmentOp)
	commitmentProof := commitmentOp.GetProof()
	return commitmentOp, commitmentProof.GetExist(), nil
}

// GetFraudulentBlockHeight returns the height of the block in which the fraud occurred
func (f *FraudProof) GetFraudulentBlockHeight() int64 {
	// Since the block height is the last committed block, the next block height is the fraudulent block height
	return f.BlockHeight + 1
}

func (f *FraudProof) GetModules() []string {
	keys := make([]string, 0, len(f.stateWitness))
	for k := range f.stateWitness {
		keys = append(keys, k)
	}
	return keys
}

// GetDeepIAVLTrees returns a map from storeKey to IAVL Deep Subtrees which have witness data and
// initial root hash initialized from fraud proof
func (f *FraudProof) GetDeepIAVLTrees() (map[string]*iavl.DeepSubTree, error) {
	storeKeyToIAVLTree := make(map[string]*iavl.DeepSubTree)
	for storeKey, stateWitness := range f.stateWitness {
		dst := iavl.NewDeepSubTree(db.NewMemDB(), 100, false, f.BlockHeight)
		iavlWitnessData := make([]iavl.WitnessData, 0)
		for _, witnessData := range stateWitness.WitnessData {
			existenceProofs, err := convertToExistenceProofs(witnessData.Proofs)
			if err != nil {
				return nil, err
			}
			iavlWitnessData = append(
				iavlWitnessData,
				iavl.WitnessData{
					Operation: witnessData.Operation,
					Key:       witnessData.Key,
					Value:     witnessData.Value,
					Proofs:    existenceProofs,
				},
			)
			dst.SetWitnessData(iavlWitnessData)
		}
		dst.SetInitialRootHash(stateWitness.RootHash)
		storeKeyToIAVLTree[storeKey] = dst
	}
	return storeKeyToIAVLTree, nil
}

// Returns true only if only one of the three pointers is nil
func (f *FraudProof) checkFraudulentStateTransition() bool {
	if f.FraudulentBeginBlock != nil {
		return f.FraudulentDeliverTx == nil && f.FraudulentEndBlock == nil
	}
	if f.FraudulentDeliverTx != nil {
		return f.FraudulentEndBlock == nil
	}
	return f.FraudulentEndBlock != nil
}

// ValidateBasic performs fraud proof verification on a store and substore level
func (f *FraudProof) ValidateBasic() (bool, error) {
	if !f.checkFraudulentStateTransition() {
		return false, ErrMoreThanOneBlockTypeUsed
	}
	for storeKey, stateWitness := range f.stateWitness {
		// Fraudproof verification on a store level
		proofOp := stateWitness.Proof
		proof, err := types.CommitmentOpDecoder(proofOp)
		if err != nil {
			return false, err
		}
		if !bytes.Equal(proof.GetKey(), []byte(storeKey)) {
			return false, fmt.Errorf("got storeKey: %s, expected: %s", string(proof.GetKey()), storeKey)
		}
		appHash, err := proof.Run([][]byte{stateWitness.RootHash})
		if err != nil {
			return false, err
		}
		if !bytes.Equal(appHash[0], f.PreStateAppHash) {
			return false, fmt.Errorf("got appHash: %s, expected: %s", string(f.PreStateAppHash), string(f.PreStateAppHash))
		}

		// Fraudproof verification on a substore level
		// Note: We can only verify the first witness in this witnessData
		// with current root hash. Other proofs are verified in the IAVL tree.
		if len(stateWitness.WitnessData) > 0 {
			witness := stateWitness.WitnessData[0]
			for _, proofOp := range witness.Proofs {
				op, existenceProof, err := getExistenceProof(*proofOp)
				if err != nil {
					return false, err
				}
				verified := ics23.VerifyMembership(op.Spec, stateWitness.RootHash, op.Proof, op.Key, existenceProof.Value)
				if !verified {
					return false, fmt.Errorf("existence proof verification failed, expected rootHash: %s, key: %s, value: %s for storeKey: %s", string(stateWitness.RootHash), string(op.Key), string(existenceProof.Value), storeKey)
				}
			}
		}
	}
	return true, nil
}

func toABCI(operation iavl.Operation) (abci.Operation, error) {
	switch operation {
	case iavl.WriteOp:
		return abci.Operation_write, nil
	case iavl.ReadOp:
		return abci.Operation_read, nil
	case iavl.DeleteOp:
		return abci.Operation_delete, nil
	default:
		return -1, fmt.Errorf("unsupported opearation: %s", operation)
	}
}

func fromABCI(operation abci.Operation) (iavl.Operation, error) {
	switch operation {
	case abci.Operation_write:
		return iavl.WriteOp, nil
	case abci.Operation_read:
		return iavl.ReadOp, nil
	case abci.Operation_delete:
		return iavl.DeleteOp, nil
	default:
		return iavl.Operation("unknown"), fmt.Errorf("unsupported opearation: %s", operation.String())
	}
}

func (f *FraudProof) toABCI() (*abci.FraudProof, error) {
	abciStateWitness := make(map[string]*abci.StateWitness)
	for storeKey, stateWitness := range f.stateWitness {
		abciWitnessData := make([]*abci.WitnessData, 0, len(stateWitness.WitnessData))
		for _, witnessData := range stateWitness.WitnessData {
			abciOperation, err := toABCI(witnessData.Operation)
			if err != nil {
				return nil, err
			}
			abciWitness := abci.WitnessData{
				Operation: abciOperation,
				Key:       witnessData.Key,
				Value:     witnessData.Value,
				Proofs:    witnessData.Proofs,
			}
			abciWitnessData = append(abciWitnessData, &abciWitness)
		}
		proof := stateWitness.Proof
		abciStateWitness[storeKey] = &abci.StateWitness{
			Proof:       &proof,
			RootHash:    stateWitness.RootHash,
			WitnessData: abciWitnessData,
		}
	}
	return &abci.FraudProof{
		BlockHeight:          f.BlockHeight,
		PreStateAppHash:      f.PreStateAppHash,
		ExpectedValidAppHash: f.ExpectedValidAppHash,
		StateWitness:         abciStateWitness,
		FraudulentBeginBlock: f.FraudulentBeginBlock,
		FraudulentDeliverTx:  f.FraudulentDeliverTx,
		FraudulentEndBlock:   f.FraudulentEndBlock,
	}, nil
}

func (f *FraudProof) FromABCI(abciFraudProof abci.FraudProof) error {
	stateWitness := make(map[string]StateWitness)
	for storeKey, abciStateWitness := range abciFraudProof.StateWitness {
		witnessData := make([]*WitnessData, 0, len(abciStateWitness.WitnessData))
		for _, abciWitnessData := range abciStateWitness.WitnessData {
			iavlOperation, err := fromABCI(abciWitnessData.Operation)
			if err != nil {
				return err
			}
			witness := WitnessData{
				Operation: iavlOperation,
				Key:       abciWitnessData.Key,
				Value:     abciWitnessData.Value,
				Proofs:    abciWitnessData.Proofs,
			}
			witnessData = append(witnessData, &witness)
		}
		stateWitness[storeKey] = StateWitness{
			Proof:       *abciStateWitness.Proof,
			RootHash:    abciStateWitness.RootHash,
			WitnessData: witnessData,
		}
	}
	f.BlockHeight = abciFraudProof.BlockHeight
	f.PreStateAppHash = abciFraudProof.PreStateAppHash
	f.ExpectedValidAppHash = abciFraudProof.ExpectedValidAppHash
	f.stateWitness = stateWitness
	f.FraudulentBeginBlock = abciFraudProof.FraudulentBeginBlock
	f.FraudulentDeliverTx = abciFraudProof.FraudulentDeliverTx
	f.FraudulentEndBlock = abciFraudProof.FraudulentEndBlock
	return nil
}
