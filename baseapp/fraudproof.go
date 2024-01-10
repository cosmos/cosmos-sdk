package baseapp

import (
	"bytes"
	"errors"
	"fmt"

	ics23 "github.com/confio/ics23/go"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/iavl"
	abci "github.com/tendermint/tendermint/abci/types"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
	db "github.com/tendermint/tm-db"
)

var ErrMoreThanOneBlockTypeUsed = errors.New("fraudProof has more than one type of fradulent state transitions marked nil")

// Represents a single-round fraudProof
type FraudProof struct {
	// The block height to load state of
	blockHeight int64

	// TODO: Add Proof that appHash is inside merklized ISRs in block header at block height

	preStateAppHash      []byte
	expectedValidAppHash []byte
	// A map from module name to state witness
	stateWitness map[string]StateWitness

	// Fraudulent state transition has to be one of these
	// Only one have of these three can be non-nil
	fraudulentBeginBlock *abci.RequestBeginBlock
	fraudulentDeliverTx  *abci.RequestDeliverTx
	fraudulentEndBlock   *abci.RequestEndBlock

	// TODO: Add Proof that fraudulent state transition is inside merkelizied transactions in block header
}

// State witness with a list of all witness data
type StateWitness struct {
	// store level proof
	Proof    tmcrypto.ProofOp
	RootHash []byte
	// List of witness data
	WitnessData []*WitnessData
}

// Witness data represents a trace operation along with inclusion proofs required for said operation
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

func (fraudProof *FraudProof) getModules() []string {
	keys := make([]string, 0, len(fraudProof.stateWitness))
	for k := range fraudProof.stateWitness {
		keys = append(keys, k)
	}
	return keys
}

func (app *BaseApp) executeNonFraudulentTransactions(req abci.RequestGenerateFraudProof, isDeliverTxFraudulent bool) {
	numNonFraudulentRequests := len(req.DeliverTxRequests)
	if isDeliverTxFraudulent {
		numNonFraudulentRequests--
	}
	nonFraudulentRequests := req.DeliverTxRequests[:numNonFraudulentRequests]
	for _, deliverTxRequest := range nonFraudulentRequests {
		app.DeliverTx(*deliverTxRequest)
	}
}

// Generate a fraudproof for an app with the given trace buffers
func (app *BaseApp) getFraudProof(storeKeyToWitnessData map[string][]iavl.WitnessData) (FraudProof, error) {
	fraudProof := FraudProof{}
	fraudProof.stateWitness = make(map[string]StateWitness)
	fraudProof.blockHeight = app.LastBlockHeight()
	cms := app.cms.(*rootmulti.Store)

	appHash, err := cms.GetAppHash()
	if err != nil {
		return FraudProof{}, err
	}
	fraudProof.preStateAppHash = appHash
	for storeKeyName := range storeKeyToWitnessData {
		iavlStore, err := cms.GetIAVLStore(storeKeyName)
		if err != nil {
			return FraudProof{}, err
		}
		rootHash, err := iavlStore.Root()
		if err != nil {
			return FraudProof{}, err
		}
		if rootHash == nil {
			continue
		}
		proof, err := cms.GetStoreProof(storeKeyName)
		if err != nil {
			return FraudProof{}, err
		}
		iavlWitnessData := storeKeyToWitnessData[storeKeyName]
		stateWitness := StateWitness{
			Proof:       *proof,
			RootHash:    rootHash,
			WitnessData: make([]*WitnessData, 0, len(iavlWitnessData)),
		}
		populateStateWitness(&stateWitness, iavlWitnessData)
		fraudProof.stateWitness[storeKeyName] = stateWitness
	}

	return fraudProof, nil
}

// populates the given state witness using the given witness data
func populateStateWitness(stateWitness *StateWitness, iavlWitnessData []iavl.WitnessData) {
	for _, iavlTraceOp := range iavlWitnessData {
		proofOps := convertToProofOps(iavlTraceOp.Proofs)
		witnessData := WitnessData{
			Operation: iavlTraceOp.Operation,
			Key:       iavlTraceOp.Key,
			Value:     iavlTraceOp.Value,
			Proofs:    proofOps,
		}
		stateWitness.WitnessData = append(stateWitness.WitnessData, &witnessData)
	}
}

// Returns a map from storeKey to IAVL Deep Subtrees which have witness data and
// initial root hash initialized from fraud proof
func (fraudProof *FraudProof) getDeepIAVLTrees() (map[string]*iavl.DeepSubTree, error) {
	storeKeyToIAVLTree := make(map[string]*iavl.DeepSubTree)
	for storeKey, stateWitness := range fraudProof.stateWitness {
		dst := iavl.NewDeepSubTree(db.NewMemDB(), 100, false, fraudProof.blockHeight)
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
func (fraudProof *FraudProof) checkFraudulentStateTransition() bool {
	if fraudProof.fraudulentBeginBlock != nil {
		return fraudProof.fraudulentDeliverTx == nil && fraudProof.fraudulentEndBlock == nil
	}
	if fraudProof.fraudulentDeliverTx != nil {
		return fraudProof.fraudulentEndBlock == nil
	}
	return fraudProof.fraudulentEndBlock != nil
}

// Performs fraud proof verification on a store and substore level
func (fraudProof *FraudProof) verifyFraudProof() (bool, error) {
	if !fraudProof.checkFraudulentStateTransition() {
		return false, ErrMoreThanOneBlockTypeUsed
	}
	for storeKey, stateWitness := range fraudProof.stateWitness {
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
		if !bytes.Equal(appHash[0], fraudProof.preStateAppHash) {
			return false, fmt.Errorf("got appHash: %s, expected: %s", string(fraudProof.preStateAppHash), string(fraudProof.preStateAppHash))
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

func (fraudProof *FraudProof) toABCI() (*abci.FraudProof, error) {
	abciStateWitness := make(map[string]*abci.StateWitness)
	for storeKey, stateWitness := range fraudProof.stateWitness {
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
		BlockHeight:          fraudProof.blockHeight,
		PreStateAppHash:      fraudProof.preStateAppHash,
		ExpectedValidAppHash: fraudProof.expectedValidAppHash,
		StateWitness:         abciStateWitness,
		FraudulentBeginBlock: fraudProof.fraudulentBeginBlock,
		FraudulentDeliverTx:  fraudProof.fraudulentDeliverTx,
		FraudulentEndBlock:   fraudProof.fraudulentEndBlock,
	}, nil
}

func (fraudProof *FraudProof) fromABCI(abciFraudProof abci.FraudProof) error {
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
	fraudProof.blockHeight = abciFraudProof.BlockHeight
	fraudProof.preStateAppHash = abciFraudProof.PreStateAppHash
	fraudProof.expectedValidAppHash = abciFraudProof.ExpectedValidAppHash
	fraudProof.stateWitness = stateWitness
	fraudProof.fraudulentBeginBlock = abciFraudProof.FraudulentBeginBlock
	fraudProof.fraudulentDeliverTx = abciFraudProof.FraudulentDeliverTx
	fraudProof.fraudulentEndBlock = abciFraudProof.FraudulentEndBlock
	return nil
}
