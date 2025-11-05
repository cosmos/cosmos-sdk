package simapp

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/telemetry"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/iavlx"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// BlockExecutionBenchConfig holds configuration for block execution benchmarks
type BlockExecutionBenchConfig struct {
	NumAccounts    int
	TxsPerBlock    int
	NumBlocks      int
	DBBackend      string
	IAVLXOptions   *iavlx.Options
	CommitBlocks   bool
	SendAmount     int64
	InitialBalance int64
}

func TestMain(m *testing.M) {
	telemetry.TestingMain(m, nil)
}

// DefaultBlockExecutionBenchConfig returns a default configuration
func DefaultBlockExecutionBenchConfig() BlockExecutionBenchConfig {
	return BlockExecutionBenchConfig{
		NumAccounts:    65_000,
		TxsPerBlock:    5000,
		NumBlocks:      150,
		DBBackend:      "memdb",
		IAVLXOptions:   nil,
		CommitBlocks:   true,
		SendAmount:     1,
		InitialBalance: 1_000_000_000,
	}
}

type accountInfo struct {
	privKey cryptotypes.PrivKey
	address sdk.AccAddress
	accNum  uint64
	seqNum  uint64
}

// BenchmarkBlockExecution benchmarks block execution with pre-built transactions
func BenchmarkBlockExecution(b *testing.B) {
	config := getBenchConfigFromEnv()
	runBlockExecutionBenchmark(b, config)
}

// BenchmarkBlockExecutionMemDB runs the benchmark with memdb
func BenchmarkBlockExecutionPebbleDB(b *testing.B) {
	config := DefaultBlockExecutionBenchConfig()
	config.DBBackend = string(dbm.PebbleDBBackend)
	config.CommitBlocks = true
	runBlockExecutionBenchmark(b, config)
}

// BenchmarkBlockExecutionIAVLX runs the benchmark with IAVLX
func BenchmarkBlockExecutionIAVLX(b *testing.B) {
	config := DefaultBlockExecutionBenchConfig()
	config.DBBackend = "iavlx"
	var iavlxOpts iavlx.Options
	iavlxOptsBz := []byte(`{"zero_copy":true,"evict_depth":20,"write_wal":true,"wal_sync_buffer":256, "fsync_interval":100,"compact_wal":true,"disable_compaction":false,"compaction_orphan_ratio":0.75,"compaction_orphan_age":10,"retain_versions":3,"min_compaction_seconds":60,"changeset_max_target":1073741824,"compaction_max_target":4294967295,"compact_after_versions":1000,"reader_update_interval":256}`)
	err := json.Unmarshal(iavlxOptsBz, &iavlxOpts)
	require.NoError(b, err)
	config.IAVLXOptions = &iavlxOpts
	config.CommitBlocks = true
	runBlockExecutionBenchmark(b, config)
}

func runBlockExecutionBenchmark(b *testing.B, config BlockExecutionBenchConfig) {
	b.ReportAllocs()

	b.Logf("Benchmark Configuration:")
	b.Logf("  Accounts: %d", config.NumAccounts)
	b.Logf("  Txs/Block: %d", config.TxsPerBlock)
	b.Logf("  Num Blocks: %d", config.NumBlocks)
	b.Logf("  DB Backend: %s", config.DBBackend)
	b.Logf("  Commit Blocks: %v", config.CommitBlocks)
	b.Logf("  Total Txs: %d", config.TxsPerBlock*config.NumBlocks)

	// Setup database
	var db dbm.DB
	var err error
	dir := b.TempDir()

	if config.DBBackend == "iavlx" {
		db, err = dbm.NewDB("application", "goleveldb", dir)
		require.NoError(b, err)
	} else {
		db, err = dbm.NewDB("application", dbm.BackendType(config.DBBackend), dir)
		require.NoError(b, err)
	}
	defer db.Close()

	homeDir := filepath.Join(dir, "home")
	err = os.MkdirAll(homeDir, 0o755)
	require.NoError(b, err)

	// Create startup configuration
	startupConfig := simtestutil.DefaultStartUpConfig()
	startupConfig.DB = db
	startupConfig.AtGenesis = true // Stay at genesis, don't finalize block 1 yet

	// Setup IAVLX if needed
	if config.DBBackend == "iavlx" {
		if config.IAVLXOptions == nil {
			config.IAVLXOptions = &iavlx.Options{
				WriteWAL:       false,
				CompactWAL:     false,
				EvictDepth:     255,
				RetainVersions: 1,
			}
		}

		startupConfig.BaseAppOption = func(app *baseapp.BaseApp) {
			iavlxDir := filepath.Join(homeDir, "data", "iavlx")
			iavlxDB, err := iavlx.LoadDB(iavlxDir, config.IAVLXOptions, log.NewNopLogger())
			if err != nil {
				b.Fatalf("Failed to load IAVLX DB: %v", err)
			}
			app.SetCMS(iavlxDB)
		}
	}

	// Create genesis accounts with funding
	fundingAmount := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, config.InitialBalance))
	genesisAccounts := make([]simtestutil.GenesisAccount, config.NumAccounts)
	accounts := make([]accountInfo, config.NumAccounts)

	for i := 0; i < config.NumAccounts; i++ {
		privKey := secp256k1.GenPrivKey()
		addr := sdk.AccAddress(privKey.PubKey().Address())
		baseAcc := authtypes.NewBaseAccount(addr, privKey.PubKey(), uint64(i), 0)

		genesisAccounts[i] = simtestutil.GenesisAccount{
			GenesisAccount: baseAcc,
			Coins:          fundingAmount,
		}

		accounts[i] = accountInfo{
			privKey: privKey,
			address: addr,
			accNum:  uint64(i),
			seqNum:  0,
		}
		if i != 0 && i%5_000 == 0 {
			b.Logf("Built accounts %d", i)
		}
	}

	startupConfig.GenesisAccounts = genesisAccounts

	// Create the app using depinject with simtestutil helpers
	var app *runtime.App
	app, err = simtestutil.SetupWithConfiguration(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.ParamsModule(),
				configurator.AuthModule(),
				configurator.StakingModule(),
				configurator.TxModule(),
				configurator.ConsensusModule(),
				configurator.BankModule(),
			),
			depinject.Supply(log.NewNopLogger()),
		),
		startupConfig,
	)
	require.NoError(b, err)
	require.NotNil(b, app)

	// Pre-build all transactions for all blocks
	b.Log("Pre-building transactions...")
	allBlockTxs := make([][][]byte, config.NumBlocks)
	txConfig := moduletestutil.MakeTestTxConfig()
	txEncoder := txConfig.TxEncoder()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	sendAmount := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, config.SendAmount))

	for blockIdx := 0; blockIdx < config.NumBlocks; blockIdx++ {
		blockTxs := make([][]byte, config.TxsPerBlock)

		for txIdx := 0; txIdx < config.TxsPerBlock; txIdx++ {
			// Select sender and recipient (ensure they're different)
			senderIdx := r.Intn(config.NumAccounts)
			recipientIdx := (senderIdx + 1 + r.Intn(config.NumAccounts-1)) % config.NumAccounts

			sender := accounts[senderIdx]
			recipient := accounts[recipientIdx]

			// Create MsgSend
			msg := banktypes.NewMsgSend(sender.address, recipient.address, sendAmount)

			// Build and sign transaction
			tx, err := simtestutil.GenSignedMockTx(
				r,
				txConfig,
				[]sdk.Msg{msg},
				sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)),
				simtestutil.DefaultGenTxGas,
				"",
				[]uint64{sender.accNum},
				[]uint64{sender.seqNum},
				sender.privKey,
			)
			require.NoError(b, err)

			// Encode transaction
			txBytes, err := txEncoder(tx)
			require.NoError(b, err)

			blockTxs[txIdx] = txBytes

			// Update sequence number for next transaction
			accounts[senderIdx].seqNum++
		}

		b.Logf("Block %d built", blockIdx)
		allBlockTxs[blockIdx] = blockTxs
	}

	b.Log("Finished pre-building transactions")

	// Reset timer to exclude setup time
	b.ResetTimer()

	// Execute blocks (start from height 1 after genesis)
	height := int64(1)
	for blockIdx := range config.NumBlocks {
		_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Txs:    allBlockTxs[blockIdx],
			Time:   time.Now(),
		})
		if err != nil {
			b.Fatalf("FinalizeBlock failed at height %d: %v", height, err)
		}

		if config.CommitBlocks {
			_, err = app.Commit()
			if err != nil {
				b.Fatalf("Commit failed at height %d: %v", height, err)
			}
		}

		height++
	}

	b.StopTimer()

	// Report statistics
	totalTxs := config.TxsPerBlock * config.NumBlocks
	b.ReportMetric(float64(totalTxs)/b.Elapsed().Seconds(), "txs/sec")
	b.ReportMetric(float64(config.NumBlocks)/b.Elapsed().Seconds(), "blocks/sec")
	b.ReportMetric(float64(totalTxs), "total_txs")
}

func getBenchConfigFromEnv() BlockExecutionBenchConfig {
	config := DefaultBlockExecutionBenchConfig()

	if val := os.Getenv("BENCH_NUM_ACCOUNTS"); val != "" {
		fmt.Sscanf(val, "%d", &config.NumAccounts)
	}

	if val := os.Getenv("BENCH_TXS_PER_BLOCK"); val != "" {
		fmt.Sscanf(val, "%d", &config.TxsPerBlock)
	}

	if val := os.Getenv("BENCH_NUM_BLOCKS"); val != "" {
		fmt.Sscanf(val, "%d", &config.NumBlocks)
	}

	if val := os.Getenv("BENCH_DB_BACKEND"); val != "" {
		config.DBBackend = val
	}

	if val := os.Getenv("BENCH_COMMIT_BLOCKS"); val != "" {
		config.CommitBlocks = val == "true"
	}

	// Check for IAVLX configuration
	if iavlxOpts := os.Getenv("IAVLX"); iavlxOpts != "" {
		config.DBBackend = "iavlx"
		var opts iavlx.Options
		err := json.Unmarshal([]byte(iavlxOpts), &opts)
		if err == nil {
			config.IAVLXOptions = &opts
		}
	}

	return config
}
