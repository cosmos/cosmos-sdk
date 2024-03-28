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

var ErrNotExactlyOneTransitionTypePresent = errors.New("fraud proof has not exactly one type of fraudulent state transitions marked nil")

// FraudProof represents a single-round fraudProof
type FraudProof struct {
	// The block height to load state of, aka the last committed block. Note: this diverges from the ADR
	BlockHeight int64

	PreStateAppHash      []byte
	ExpectedValidAppHash []byte
	// A map from module name to state witness
	moduleToWitness map[string]StateWitness

	// Fraudulent state transition has to be one of these
	// Only one of these three can be non-nil
	FraudulentBeginBlock *abci.RequestBeginBlock
	FraudulentDeliverTx  *abci.RequestDeliverTx
	FraudulentEndBlock   *abci.RequestEndBlock

	// TODO(danwt): see celestia todos https://github.com/celestiaorg/cosmos-sdk/compare/release/v0.46.x-celestia...rollkit:cosmos-sdk-old:manav/fraudproof_iavl_prototype#diff-b5f489a3fbc869bd5596de0eea860d2c9e44bcc3793be9b86bb24cc78460f9aaR23-R36
	// TODO: (?) Add Proof that appHash is inside merklized ISRs in block header at block height
	// TODO: (?) Add Proof that fraudulent state transition is inside merkelizied transactions in block header
}

// StateWitness with a list of all witness data, for a module
type StateWitness struct {
	// store level proof (proof of the substore belonging to the global store)
	Proof       tmcrypto.ProofOp
	RootHash    []byte
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

// convertToExistenceProofs converts a slice of ProofOps to a slice of ExistenceProofs
// it's purely a type conversion, no logic happens
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

// getExistenceProof converts a tendermint ProofOp to a commitment and an existence proof
// it's purely a type conversion, no logic happens
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
	return f.GetStateWitnessKeys()
}

func (f *FraudProof) GetStateWitnessKeys() []string {
	keys := make([]string, 0, len(f.moduleToWitness))
	for k := range f.moduleToWitness {
		keys = append(keys, k)
	}
	return keys
}

// GetModuleToDeepIAVLTree returns a map from module store keys to IAVL Deep Subtrees
// which have witness data an initial root hash initialized from fraud proof
func (f *FraudProof) GetModuleToDeepIAVLTree() (map[string]*iavl.DeepSubTree, error) {
	ret := make(map[string]*iavl.DeepSubTree)
	for moduleStoreKey, w := range f.moduleToWitness {
		treeVersion := f.BlockHeight // TODO(danwt): explain
		cacheSize := 100             // Copied from celestia
		// create an empty tree
		tree := iavl.NewDeepSubTree(db.NewMemDB(), cacheSize, false, treeVersion)

		// convert the sdk witness data to the iavl witness data format
		// this populates the tree for each operation we are going to do
		iavlWitnessData := make([]iavl.WitnessData, 0)
		for _, witnessData := range w.WitnessData {
			existenceProofs, err := convertToExistenceProofs(witnessData.Proofs)
			if err != nil {
				return nil, fmt.Errorf("convert to existence proofs: %w", err)
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
			tree.SetWitnessData(iavlWitnessData)
		}
		tree.SetInitialRootHash(w.RootHash)
		ret[moduleStoreKey] = tree
	}
	return ret, nil
}

func (f *FraudProof) exactlyOneTransition() bool {
	cnt := 0
	if f.FraudulentBeginBlock != nil {
		cnt++
	}
	if f.FraudulentDeliverTx != nil {
		cnt++
	}
	if f.FraudulentEndBlock != nil {
		cnt++
	}
	return cnt == 1
}

// ValidateBasic checks that the fraud proof is well-formed
// based on https://github.com/celestiaorg/cosmos-sdk/compare/release/v0.46.x-celestia...rollkit:cosmos-sdk-old:manav/fraudproof_iavl_prototype#diff-b5f489a3fbc869bd5596de0eea860d2c9e44bcc3793be9b86bb24cc78460f9aaR147-R187
func (f *FraudProof) ValidateBasic() error {
	if !f.exactlyOneTransition() {
		return ErrNotExactlyOneTransitionTypePresent
	}

	for storeKey, stateWitness := range f.moduleToWitness {
		proofOp := stateWitness.Proof
		proof, err := types.CommitmentOpDecoder(proofOp)
		if err != nil {
			return err
		}

		// Each proof must correspond to the correct key
		if !bytes.Equal(proof.GetKey(), []byte(storeKey)) {
			return fmt.Errorf("got storeKey: %s, expected: %s", string(proof.GetKey()), storeKey)
		}

		// Each substore must prove to the correct app hash
		appHash, err := proof.Run([][]byte{stateWitness.RootHash})
		if err != nil {
			return err
		}
		if !bytes.Equal(appHash[0], f.PreStateAppHash) {
			return fmt.Errorf("got appHash: %s, expected: %s", string(f.PreStateAppHash), string(f.PreStateAppHash))
		}

		// Now check inside the substore proofs
		// Note: We can only verify the first witness in this witnessData
		// with current root hash. Other proofs are verified in the IAVL tree. TODO(danwt): explain why
		if 0 < len(stateWitness.WitnessData) {
			witness := stateWitness.WitnessData[0]
			for _, proofOp := range witness.Proofs {
				op, existenceProof, err := getExistenceProof(*proofOp)
				if err != nil {
					return err
				}
				verified := ics23.VerifyMembership(op.Spec, stateWitness.RootHash, op.Proof, op.Key, existenceProof.Value)
				if !verified {
					return fmt.Errorf(
						"existence proof verification failed, expected rootHash: %s, key: %s, value: %s for storeKey: %s",
						string(stateWitness.RootHash),
						string(op.Key),
						string(existenceProof.Value),
						storeKey,
					)
				}
			}
		}
	}
	return nil
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
	for storeKey, stateWitness := range f.moduleToWitness {
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
	f.moduleToWitness = stateWitness
	f.FraudulentBeginBlock = abciFraudProof.FraudulentBeginBlock
	f.FraudulentDeliverTx = abciFraudProof.FraudulentDeliverTx
	f.FraudulentEndBlock = abciFraudProof.FraudulentEndBlock
	return nil
}
