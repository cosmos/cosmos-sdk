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
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/proxy"
	tmsm "github.com/tendermint/tendermint/state"
	tm "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/server"
)

var (
	rootDir string
)

var rootCmd = &cobra.Command{
	Use:   "gaiareplay",
	Short: "Replay gaia transactions",
	Run: func(cmd *cobra.Command, args []string) {
		run(rootDir)
	},
}

func init() {
	// cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&rootDir, "root", "r", "root dir")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(rootDir string) {

	// Copy the rootDir to a new directory, to preserve the old one.
	oldRootDir := rootDir
	rootDir = oldRootDir + "_replay"
	if cmn.FileExists(rootDir) {
		cmn.Exit(fmt.Sprintf("temporary copy dir %v already exists", rootDir))
	}
	err := cpm.Copy(oldRootDir, rootDir)
	if err != nil {
		panic(err)
	}

	configDir := filepath.Join(rootDir, "config")
	dataDir := filepath.Join(rootDir, "data")
	ctx := server.NewDefaultContext()

	// App DB
	// appDB := dbm.NewMemDB()
	appDB, err := dbm.NewGoLevelDB("application", dataDir)
	if err != nil {
		panic(err)
	}

	// TM DB
	// tmDB := dbm.NewMemDB()
	tmDB, err := dbm.NewGoLevelDB("state", dataDir)
	if err != nil {
		panic(err)
	}

	// Blockchain DB
	bcDB, err := dbm.NewGoLevelDB("blockstore", dataDir)
	if err != nil {
		panic(err)
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
		panic(err)
	}

	// Application
	myapp := app.NewGaiaApp(
		ctx.Logger, appDB, traceStoreWriter,
		baseapp.SetPruning("everything"), // nothing
	)

	// Genesis
	var genDocPath = filepath.Join(configDir, "genesis.json")
	genDoc, err := tm.GenesisDocFromFile(genDocPath)
	if err != nil {
		panic(err)
	}
	genState, err := tmsm.MakeGenesisState(genDoc)
	if err != nil {
		panic(err)
	}
	// tmsm.SaveState(tmDB, genState)

	cc := proxy.NewLocalClientCreator(myapp)
	proxyApp := proxy.NewAppConns(cc)
	err = proxyApp.Start()
	if err != nil {
		panic(err)
	}
	defer proxyApp.Stop()

	// Send InitChain msg
	validators := tm.TM2PB.ValidatorUpdates(genState.Validators)
	csParams := tm.TM2PB.ConsensusParams(genDoc.ConsensusParams)
	req := abci.RequestInitChain{
		Time:            genDoc.GenesisTime,
		ChainId:         genDoc.ChainID,
		ConsensusParams: csParams,
		Validators:      validators,
		AppStateBytes:   genDoc.AppState,
	}
	_, err = proxyApp.Consensus().InitChainSync(req)
	if err != nil {
		panic(err)
	}

	// Create executor
	blockExec := tmsm.NewBlockExecutor(tmDB, ctx.Logger, proxyApp.Consensus(),
		tmsm.MockMempool{}, tmsm.MockEvidencePool{})

	// Create block store
	blockStore := bcm.NewBlockStore(bcDB)

	// Update this state.
	state := genState
	tz := []time.Duration{0, 0, 0}
	for i := 1; i < 1e10; i++ {

		t1 := time.Now()

		// Apply block
		fmt.Printf("loading and applying block %d\n", i)
		blockmeta := blockStore.LoadBlockMeta(int64(i))
		if blockmeta == nil {
			panic(fmt.Sprintf("couldn't find block meta %d", i))
		}
		block := blockStore.LoadBlock(int64(i))
		if block == nil {
			panic(fmt.Sprintf("couldn't find block %d", i))
		}

		t2 := time.Now()

		state, err = blockExec.ApplyBlock(state, blockmeta.BlockID, block)
		if err != nil {
			panic(err)
		}

		t3 := time.Now()
		tz[0] += t2.Sub(t1)
		tz[1] += t3.Sub(t2)

		fmt.Printf("new app hash: %X\n", state.AppHash)
		fmt.Println(tz)
	}

}
