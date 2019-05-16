package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	cpm "github.com/otiai10/copy"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"
	bcm "github.com/tendermint/tendermint/blockchain"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/proxy"
	tmsm "github.com/tendermint/tendermint/state"
	tm "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func replayCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "replay <root-dir>",
		Short: "Replay gaia transactions",
		RunE: func(_ *cobra.Command, args []string) error {
			return replayTxs(args[0])
		},
		Args: cobra.ExactArgs(1),
	}
}

func replayTxs(rootDir string) error {

	if false {
		// Copy the rootDir to a new directory, to preserve the old one.
		fmt.Fprintln(os.Stderr, "Copying rootdir over")
		oldRootDir := rootDir
		rootDir = oldRootDir + "_replay"
		if cmn.FileExists(rootDir) {
			cmn.Exit(fmt.Sprintf("temporary copy dir %v already exists", rootDir))
		}
		if err := cpm.Copy(oldRootDir, rootDir); err != nil {
			return err
		}
	}

	configDir := filepath.Join(rootDir, "config")
	dataDir := filepath.Join(rootDir, "data")
	ctx := server.NewDefaultContext()

	// App DB
	// appDB := dbm.NewMemDB()
	fmt.Fprintln(os.Stderr, "Opening app database")
	appDB, err := sdk.NewLevelDB("application", dataDir)
	if err != nil {
		return err
	}

	// TM DB
	// tmDB := dbm.NewMemDB()
	fmt.Fprintln(os.Stderr, "Opening tendermint state database")
	tmDB, err := sdk.NewLevelDB("state", dataDir)
	if err != nil {
		return err
	}

	// Blockchain DB
	fmt.Fprintln(os.Stderr, "Opening blockstore database")
	bcDB, err := sdk.NewLevelDB("blockstore", dataDir)
	if err != nil {
		return err
	}

	// TraceStore
	var traceStoreWriter io.Writer
	var traceStoreDir = filepath.Join(dataDir, "trace.log")
	traceStoreWriter, err = os.OpenFile(
		traceStoreDir,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0666,
	)
	if err != nil {
		return err
	}

	// Application
	fmt.Fprintln(os.Stderr, "Creating application")
	myapp := app.NewGaiaApp(
		ctx.Logger, appDB, traceStoreWriter, true, uint(1),
		baseapp.SetPruning(store.PruneEverything), // nothing
	)

	// Genesis
	var genDocPath = filepath.Join(configDir, "genesis.json")
	genDoc, err := tm.GenesisDocFromFile(genDocPath)
	if err != nil {
		return err
	}
	genState, err := tmsm.MakeGenesisState(genDoc)
	if err != nil {
		return err
	}
	// tmsm.SaveState(tmDB, genState)

	cc := proxy.NewLocalClientCreator(myapp)
	proxyApp := proxy.NewAppConns(cc)
	err = proxyApp.Start()
	if err != nil {
		return err
	}
	defer func() {
		_ = proxyApp.Stop()
	}()

	state := tmsm.LoadState(tmDB)
	if state.LastBlockHeight == 0 {
		// Send InitChain msg
		fmt.Fprintln(os.Stderr, "Sending InitChain msg")
		validators := tm.TM2PB.ValidatorUpdates(genState.Validators)
		csParams := tm.TM2PB.ConsensusParams(genDoc.ConsensusParams)
		req := abci.RequestInitChain{
			Time:            genDoc.GenesisTime,
			ChainId:         genDoc.ChainID,
			ConsensusParams: csParams,
			Validators:      validators,
			AppStateBytes:   genDoc.AppState,
		}
		res, err := proxyApp.Consensus().InitChainSync(req)
		if err != nil {
			return err
		}
		newValidatorz, err := tm.PB2TM.ValidatorUpdates(res.Validators)
		if err != nil {
			return err
		}
		newValidators := tm.NewValidatorSet(newValidatorz)

		// Take the genesis state.
		state = genState
		state.Validators = newValidators
		state.NextValidators = newValidators
	}

	// Create executor
	fmt.Fprintln(os.Stderr, "Creating block executor")
	blockExec := tmsm.NewBlockExecutor(tmDB, ctx.Logger, proxyApp.Consensus(),
		tmsm.MockMempool{}, tmsm.MockEvidencePool{})

	// Create block store
	fmt.Fprintln(os.Stderr, "Creating block store")
	blockStore := bcm.NewBlockStore(bcDB)

	tz := []time.Duration{0, 0, 0}
	for i := int(state.LastBlockHeight) + 1; ; i++ {
		fmt.Fprintln(os.Stderr, "Running block ", i)
		t1 := time.Now()

		// Apply block
		fmt.Printf("loading and applying block %d\n", i)
		blockmeta := blockStore.LoadBlockMeta(int64(i))
		if blockmeta == nil {
			fmt.Printf("Couldn't find block meta %d... done?\n", i)
			return nil
		}
		block := blockStore.LoadBlock(int64(i))
		if block == nil {
			return fmt.Errorf("couldn't find block %d", i)
		}

		t2 := time.Now()

		state, err = blockExec.ApplyBlock(state, blockmeta.BlockID, block)
		if err != nil {
			return err
		}

		t3 := time.Now()
		tz[0] += t2.Sub(t1)
		tz[1] += t3.Sub(t2)

		fmt.Fprintf(os.Stderr, "new app hash: %X\n", state.AppHash)
		fmt.Fprintln(os.Stderr, tz)
	}
}
