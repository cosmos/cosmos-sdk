package baseapp

import (
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/cosmos/iavl"
	abci "github.com/tendermint/tendermint/abci/types"
)

// GenerateFraudProof implements the ABCI application interface. The BaseApp reverts to
// previous state, runs the given fraudulent state transition, and gets the traced witness data representing
// the operations that this state transition makes. It then uses this traced witness data and
// the pre-fraudulent execution state of the BaseApp to generates a Fraud Proof
// representing it. It returns this generated Fraud Proof.
func (app *BaseApp) GenerateFraudProof(req abci.RequestGenerateFraudProof) (res abci.ResponseGenerateFraudProof) {
	logger := app.logger.With("module", "fraudproof")

	// Revert app to previous state
	cms := app.cms.(*rootmulti.Store)
	cms.SetInterBlockCache(nil)
	err := cms.LoadLastVersion()
	if err != nil {
		// Happens when there is no last state to load form
		panic(err)
	}
	logger.Debug("Initial", "AppHash", hex.EncodeToString(app.GetAppHashInternal()))

	app.deliverState = nil

	// Run the set of all nonFradulent and fraudulent state transitions
	beginBlockRequest := req.BeginBlockRequest
	isBeginBlockFraudulent := req.DeliverTxRequests == nil
	isDeliverTxFraudulent := req.EndBlockRequest == nil
	if isBeginBlockFraudulent {
		cms.SetTracingEnabledAll(true)
	}

	app.BeginBlock(beginBlockRequest)
	logger.Debug("after beginBlock", "AppHash", hex.EncodeToString(app.GetAppHashInternal()))

	if !isBeginBlockFraudulent {
		// BeginBlock is not the fraudulent state transition
		app.executeNonFraudulentTransactions(req, isDeliverTxFraudulent)

		cms.SetTracingEnabledAll(true)
		// skip IncrementSequenceDecorator check in AnteHandler
		app.anteHandler = nil

		// Record the trace made by the fraudulent state transitions
		if isDeliverTxFraudulent {
			// The last DeliverTx is the fraudulent state transition
			fraudulentDeliverTx := req.DeliverTxRequests[len(req.DeliverTxRequests)-1]
			app.DeliverTx(*fraudulentDeliverTx)
			logger.Debug("after delivertx", "AppHash", hex.EncodeToString(app.GetAppHashInternal()))

		} else {
			// EndBlock is the fraudulent state transition
			app.EndBlock(*req.EndBlockRequest)
			logger.Debug("after endblock", "AppHash", hex.EncodeToString(app.GetAppHashInternal()))
		}
	}

	validAppHash := app.GetAppHash(abci.RequestGetAppHash{}).AppHash
	logger.Debug("validAppHash", "AppHash", hex.EncodeToString(validAppHash))

	storeKeyToWitnessData := cms.GetWitnessDataMap()

	//TODO: NOT sure why we need to revert, it's probably better to use cached state and discard

	// Revert app to previous state
	cms.SetInterBlockCache(nil)
	err = cms.LoadLastVersion()
	if err != nil {
		panic(err)
	}
	app.deliverState = nil
	// Fast-forward to right before fraudulent state transition occurred
	app.BeginBlock(beginBlockRequest) //TODO: Potentially move inside next if statement
	if !isBeginBlockFraudulent {
		app.executeNonFraudulentTransactions(req, isDeliverTxFraudulent)
	}

	// Export the app's current trace-filtered state into a Fraud Proof and return it
	fraudProof, err := app.getFraudProof(storeKeyToWitnessData)
	if err != nil {
		panic(err)
	}

	fraudProof.ExpectedValidAppHash = validAppHash

	switch {
	case isBeginBlockFraudulent:
		fraudProof.FraudulentBeginBlock = &beginBlockRequest
	case isDeliverTxFraudulent:
		fraudProof.FraudulentDeliverTx = req.DeliverTxRequests[len(req.DeliverTxRequests)-1]
	default:
		fraudProof.FraudulentEndBlock = req.EndBlockRequest
	}
	abciFraudProof, err := fraudProof.toABCI()
	if err != nil {
		panic(err)
	}
	res = abci.ResponseGenerateFraudProof{
		FraudProof: abciFraudProof,
	}
	return res
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
	fraudProof.BlockHeight = app.LastBlockHeight()
	cms := app.cms.(*rootmulti.Store)

	appHash, err := cms.GetAppHash()
	if err != nil {
		return FraudProof{}, err
	}
	fraudProof.PreStateAppHash = appHash

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

func (app *BaseApp) VerifyFraudProof(req abci.RequestVerifyFraudProof) (res abci.ResponseVerifyFraudProof) {
	panic("not implemented")
}
