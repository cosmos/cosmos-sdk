package baseapp

import (
	"bytes"
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/iavl"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

// GenerateFraudProof implements the ABCI application interface. The BaseApp reverts to
// previous state, runs the given fraudulent state transition, and gets the traced witness data representing
// the operations that this state transition makes. It then uses this traced witness data and
// the pre-fraudulent execution state of the BaseApp to generates a Fraud Proof
// representing it. It returns this generated Fraud Proof.
func (app *BaseApp) GenerateFraudProof(req abci.RequestGenerateFraudProof) (res abci.ResponseGenerateFraudProof) {
	// Revert app to previous state
	cms := app.cms.(*rootmulti.Store)
	cms.SetInterBlockCache(nil)
	err := cms.LoadLastVersion()
	if err != nil {
		// Happens when there is no last state to load form
		panic(err)
	}
	app.deliverState = nil

	prestateHash, err := cms.GetAppHash()
	if err != nil {
		panic(err)
	}

	app.logger.Info("AppHash", "prestateHash", hex.EncodeToString(prestateHash))

	//FIRST PHASE: tracing all block
	cms.SetTracingEnabledAll(true)
	app.BeginBlock(req.BeginBlockRequest)
	app.logger.Info("AppHash", "beginBlock", hex.EncodeToString(app.GetAppHashInternal()))

	for _, deliverTxRequest := range req.DeliverTxRequests {
		app.DeliverTx(*deliverTxRequest)
	}
	// skip IncrementSequenceDecorator check in AnteHandler
	app.anteHandler = nil
	app.logger.Info("AppHash", "delivertx", hex.EncodeToString(app.GetAppHashInternal()))

	app.EndBlock(*req.EndBlockRequest)
	validAppHash := app.GetAppHash(abci.RequestGetAppHash{}).AppHash
	app.logger.Info("AppHash", "endblock", hex.EncodeToString(validAppHash))

	storeKeyToWitnessData := cms.GetWitnessDataMap()
	// Revert app to previous state
	cms.SetInterBlockCache(nil)
	err = cms.LoadLastVersion()
	if err != nil {
		panic(err)
	}
	// app.deliverState = nil

	// Export the app's current trace-filtered state into a Fraud Proof and return it
	fraudProof, err := app.getFraudProof(storeKeyToWitnessData)
	if err != nil {
		panic(err)
	}

	fraudProof.ExpectedValidAppHash = validAppHash

	fraudProof.FraudulentBeginBlock = &req.BeginBlockRequest
	fraudProof.FraudulentDeliverTx = req.DeliverTxRequests
	fraudProof.FraudulentEndBlock = req.EndBlockRequest

	abciFraudProof, err := fraudProof.toABCI()
	if err != nil {
		panic(err)
	}
	res = abci.ResponseGenerateFraudProof{
		FraudProof: abciFraudProof,
	}
	return res
}

// VerifyFraudProof implements the ABCI application interface. It loads a fresh BaseApp using
// the given Fraud Proof, runs the given fraudulent state transition within the Fraud Proof,
// and gets the app hash representing state of the resulting BaseApp. It returns a boolean
// representing whether this app hash is equivalent to the expected app hash given.
func (app *BaseApp) VerifyFraudProof(req abci.RequestVerifyFraudProof) (res abci.ResponseVerifyFraudProof) {
	abciFraudProof := req.FraudProof
	fraudProof := FraudProof{}
	err := fraudProof.FromABCI(*abciFraudProof)
	if err != nil {
		panic(err)
	}

	// Store and subtore level verification
	success, err := fraudProof.ValidateBasic()
	if err != nil {
		panic(err)
	}

	if success {
		// State execution verification
		options := make([]func(*BaseApp), 0)
		if app.routerOpts != nil {
			for _, routerOpt := range app.routerOpts {
				options = append(options, routerOpt)
			}
		}
		cms := app.cms.(*rootmulti.Store)
		storeKeys := cms.StoreKeysByName()
		modules := fraudProof.GetModules()
		iavlStoreKeys := make([]storetypes.StoreKey, 0, len(modules))
		for _, module := range modules {
			iavlStoreKeys = append(iavlStoreKeys, storeKeys[module])
		}
		// Setup a new app from fraud proof
		appFromFraudProof, err := SetupBaseAppFromFraudProof(
			app,
			dbm.NewMemDB(),
			fraudProof,
			iavlStoreKeys,
			options...,
		)
		if err != nil {
			panic(err)
		}
		appFromFraudProof.InitChain(abci.RequestInitChain{
			InitialHeight: fraudProof.BlockHeight,
		})
		appHash := appFromFraudProof.GetAppHash(abci.RequestGetAppHash{}).AppHash

		if !bytes.Equal(fraudProof.PreStateAppHash, appHash) {
			return abci.ResponseVerifyFraudProof{
				Success: false,
			}
		}

		// Execute fraudulent state transition
		appFromFraudProof.BeginBlock(*fraudProof.FraudulentBeginBlock)
		// Need to add some dummy begin block here since its a new app
		appFromFraudProof.beginBlocker = nil
		appFromFraudProof.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: fraudProof.BlockHeight}})
		if fraudProof.FraudulentDeliverTx != nil {
			for _, deliverTxRequest := range fraudProof.FraudulentDeliverTx {
				appFromFraudProof.DeliverTx(*deliverTxRequest)
				resp := appFromFraudProof.DeliverTx(*deliverTxRequest)
				if !resp.IsOK() {
					panic(resp.Log)
				}
			}
		}
		appFromFraudProof.EndBlock(*fraudProof.FraudulentEndBlock)

		appHash = appFromFraudProof.GetAppHash(abci.RequestGetAppHash{}).AppHash
		success = bytes.Equal(appHash, req.ExpectedValidAppHash)
	}
	res = abci.ResponseVerifyFraudProof{
		Success: success,
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
	fraudProof.BlockHeight = app.LastBlockHeight() + 1 //FIXME: patch
	cms := app.cms.(*rootmulti.Store)

	appHash, err := cms.GetAppHash()
	if err != nil {
		return FraudProof{}, err
	}
	fraudProof.PreStateAppHash = appHash
	app.logger.Info("AppHash", "getFraudProof", hex.EncodeToString(appHash))

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

// set up a new baseapp from given params
func setupBaseAppFromParams(app *BaseApp, db dbm.DB, storeKeyToIAVLTree map[string]*iavl.DeepSubTree, blockHeight int64, storeKeys []storetypes.StoreKey, options ...func(*BaseApp)) (*BaseApp, error) {
	// This initial height is used in `BeginBlock` in `validateHeight`
	// options = append(options, SetInitialHeight(blockHeight))

	appName := app.Name() + "FromFraudProof"
	newApp := NewBaseApp(appName, app.logger, db, app.txDecoder, options...)

	newApp.msgServiceRouter = app.msgServiceRouter
	newApp.beginBlocker = app.beginBlocker
	newApp.endBlocker = app.endBlocker
	// stores are mounted
	newApp.MountStores(storeKeys...)
	cmsStore := newApp.cms.(*rootmulti.Store)
	for storeKey, iavlTree := range storeKeyToIAVLTree {
		cmsStore.SetDeepIAVLTree(storeKey, iavlTree)
	}
	err := newApp.LoadLatestVersion()
	return newApp, err
}

// set up a new baseapp from a fraudproof
func SetupBaseAppFromFraudProof(app *BaseApp, db dbm.DB, fraudProof FraudProof, storeKeys []storetypes.StoreKey, options ...func(*BaseApp)) (*BaseApp, error) {
	storeKeyToIAVLTree, err := fraudProof.GetDeepIAVLTrees()
	if err != nil {
		return nil, err
	}
	return setupBaseAppFromParams(app, db, storeKeyToIAVLTree, fraudProof.BlockHeight, storeKeys, options...)
}
