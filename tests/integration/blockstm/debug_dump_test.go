package blockstm_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/baseapp/txnrunner"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// debugSnapshot mirrors the JSON structure written by BlockExecutionDebug.SaveToFile.
// Defined here to avoid importing internal/blockstm from the external test module.
type debugSnapshot struct {
	BlockSize    int             `json:"block_size"`
	Transactions []debugTxnData `json:"transactions"`
}

type debugTxnData struct {
	Executions  []debugExecution  `json:"executions"`
	Validations []debugValidation `json:"validations"`
	Suspensions []debugSuspension `json:"suspensions"`
}

type debugExecution struct {
	Incarnation uint      `json:"incarnation"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
}

type debugValidation struct {
	Incarnation uint      `json:"incarnation"`
	Timestamp   time.Time `json:"timestamp"`
	Valid       bool      `json:"valid"`
	Aborted     bool      `json:"aborted"`
}

type debugSuspension struct {
	Suspend   time.Time `json:"suspend"`
	Resume    time.Time `json:"resume,omitempty"`
	BlockedBy uint32    `json:"blocked_by"`
}

// TestBlockSTM_DebugDumpOnAppHashMismatch exercises the full debug dump pipeline:
//   - A chain runs with the parallel BlockSTM executor
//   - Transactions that write to overlapping state are executed (forcing re-executions)
//   - The debug trace is persisted to disk after FinalizeBlock
//   - We load the file and inspect the data, simulating what an operator would
//     do after a CometBFT crash from an app hash mismatch
//
// In production, the debug file survives the CometBFT panic from
// assertAppHashEqualsOneFromBlock/State and is available for post-mortem analysis.
func TestBlockSTM_DebugDumpOnAppHashMismatch(t *testing.T) {
	// --- Setup ---
	debugDir := t.TempDir()

	numSenders := 30
	// All senders send to the SAME recipient, creating write conflicts
	// that force BlockSTM to abort, re-execute and re-validate transactions.
	sharedRecipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	senderAddrs := make([]sdk.AccAddress, numSenders)
	for i := range numSenders {
		senderAddrs[i] = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	}

	logger := log.NewNopLogger()
	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey, banktypes.StoreKey)
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{})
	cdc := encCfg.Codec
	txConfig := encCfg.TxConfig

	bApp := baseapp.NewBaseApp("blockstm-debug-test", logger, dbm.NewMemDB(), txConfig.TxDecoder(), baseapp.SetChainID("blockstm-debug-test"))
	bApp.MountKVStores(keys)
	bApp.SetInterfaceRegistry(encCfg.InterfaceRegistry)

	authority := authtypes.NewModuleAddress("gov")
	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		map[string][]string{minttypes.ModuleName: {authtypes.Minter}},
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)
	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		map[string]bool{accountKeeper.GetAuthority(): false},
		authority.String(),
		log.NewNopLogger(),
	)

	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)

	bApp.SetInitChainer(func(ctx sdk.Context, _ *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
		authModule.InitGenesis(ctx, cdc, authModule.DefaultGenesis(cdc))
		bankModule.InitGenesis(ctx, cdc, bankModule.DefaultGenesis(cdc))
		return &abci.ResponseInitChain{}, nil
	})
	banktypes.RegisterMsgServer(bApp.MsgServiceRouter(), bankkeeper.NewMsgServerImpl(bankKeeper))

	require.NoError(t, bApp.LoadLatestVersion())
	_, err := bApp.InitChain(&abci.RequestInitChain{ChainId: "blockstm-debug-test"})
	require.NoError(t, err)

	// FinalizeBlock to get state ready for keeper writes
	_, err = bApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: bApp.LastBlockHeight() + 1})
	require.NoError(t, err)

	// Fund all senders
	ctx := bApp.NewContext(false)
	for _, addr := range senderAddrs {
		require.NoError(t, testutil.FundAccount(ctx, bankKeeper, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10_000))))
	}

	// Commit funded state
	_, err = bApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: bApp.LastBlockHeight() + 1})
	require.NoError(t, err)
	_, err = bApp.Commit()
	require.NoError(t, err)

	// --- Enable BlockSTM with debug persistence ---
	storeKeys := make([]storetypes.StoreKey, 0, len(keys))
	for _, key := range keys {
		storeKeys = append(storeKeys, key)
	}
	runner := txnrunner.NewSTMRunner(
		txConfig.TxDecoder(),
		storeKeys,
		8,
		false,
		func(_ storetypes.MultiStore) string { return sdk.DefaultBondDenom },
	)
	bApp.SetBlockSTMTxRunner(runner)
	bApp.SetBlockSTMDebugDir(debugDir)

	// --- Build conflicting transactions ---
	// All senders send to the SAME recipient. This forces concurrent writes to
	// the recipient's balance key, causing BlockSTM to abort and re-execute
	// transactions with higher incarnation numbers.
	txBytes := make([][]byte, numSenders)
	for i := range numSenders {
		msg := banktypes.NewMsgSend(senderAddrs[i], sharedRecipient, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)))
		txBuilder := txConfig.NewTxBuilder()
		require.NoError(t, txBuilder.SetMsgs(msg))
		txBytes[i], err = txConfig.TxEncoder()(txBuilder.GetTx())
		require.NoError(t, err)
	}

	// --- Execute the block ---
	execHeight := bApp.LastBlockHeight() + 1
	blockRes, err := bApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: execHeight,
		Txs:    txBytes,
	})
	require.NoError(t, err)
	require.Len(t, blockRes.TxResults, numSenders)

	for i, result := range blockRes.TxResults {
		require.Equal(t, uint32(0), result.Code,
			"tx %d failed: code=%d log=%s", i, result.Code, result.Log)
	}

	// --- Verify the debug dump file exists ---
	debugFile := filepath.Join(debugDir, "blockstm_last_execution.json")
	require.FileExists(t, debugFile, "debug trace file should have been written after FinalizeBlock")

	// --- Load and inspect the snapshot (simulating post-crash analysis) ---
	rawData, err := os.ReadFile(debugFile)
	require.NoError(t, err)
	require.True(t, json.Valid(rawData), "debug file should contain valid JSON")

	var snap debugSnapshot
	require.NoError(t, json.Unmarshal(rawData, &snap))
	require.Equal(t, numSenders, snap.BlockSize, "snapshot block size should match number of txs")
	require.Len(t, snap.Transactions, numSenders)

	// Verify every transaction has at least one execution record with timestamps
	totalExecutions := 0
	totalValidations := 0
	totalSuspensions := 0
	maxIncarnation := uint(0)
	for i, txn := range snap.Transactions {
		require.NotEmpty(t, txn.Executions, "txn %d should have at least one execution record", i)

		for _, exec := range txn.Executions {
			require.False(t, exec.Start.IsZero(), "txn %d execution start should be set", i)
			require.False(t, exec.End.IsZero(), "txn %d execution end should be set", i)
			require.True(t, exec.End.After(exec.Start) || exec.End.Equal(exec.Start),
				"txn %d execution end should be >= start", i)
			if exec.Incarnation > maxIncarnation {
				maxIncarnation = exec.Incarnation
			}
		}

		totalExecutions += len(txn.Executions)
		totalValidations += len(txn.Validations)
		totalSuspensions += len(txn.Suspensions)
	}

	t.Logf("=== Block execution debug summary ===")
	t.Logf("  Transactions:      %d", numSenders)
	t.Logf("  Total executions:  %d (%.1fx avg per tx)", totalExecutions, float64(totalExecutions)/float64(numSenders))
	t.Logf("  Total validations: %d", totalValidations)
	t.Logf("  Total suspensions: %d", totalSuspensions)
	t.Logf("  Max incarnation:   %d", maxIncarnation)
	t.Logf("  Debug file:        %s (%d bytes)", debugFile, len(rawData))

	// With all transactions writing to the same recipient, we expect re-executions
	// (incarnation > 0) due to write conflicts.
	require.Greater(t, totalExecutions, numSenders,
		"conflicting txs should cause re-executions, so total executions > block size")
	require.Greater(t, maxIncarnation, uint(0),
		"at least one transaction should have been re-executed with incarnation > 0")
	require.Greater(t, totalValidations, 0,
		"there should be validation records")

	// Verify validation records have proper fields
	for _, txn := range snap.Transactions {
		for _, val := range txn.Validations {
			require.False(t, val.Timestamp.IsZero(), "validation timestamp should be set")
		}
	}

	// --- Log detailed trace for a sample transaction ---
	t.Logf("=== Transaction 0 detail ===")
	txn0 := snap.Transactions[0]
	for j, exec := range txn0.Executions {
		t.Logf("  exec[%d] incarnation=%d duration=%v", j, exec.Incarnation, exec.End.Sub(exec.Start))
	}
	for j, val := range txn0.Validations {
		t.Logf("  val[%d]  incarnation=%d valid=%v aborted=%v", j, val.Incarnation, val.Valid, val.Aborted)
	}
	for j, sus := range txn0.Suspensions {
		t.Logf("  sus[%d]  blocked_by=%d wait=%v", j, sus.BlockedBy, sus.Resume.Sub(sus.Suspend))
	}

	// --- Log a high-incarnation transaction showing conflict resolution ---
	for i, txn := range snap.Transactions {
		for _, exec := range txn.Executions {
			if exec.Incarnation > 1 {
				t.Logf("=== High-incarnation transaction %d ===", i)
				for j, e := range txn.Executions {
					t.Logf("  exec[%d] incarnation=%d duration=%v", j, e.Incarnation, e.End.Sub(e.Start))
				}
				for j, v := range txn.Validations {
					t.Logf("  val[%d]  incarnation=%d valid=%v aborted=%v", j, v.Incarnation, v.Valid, v.Aborted)
				}
				for j, s := range txn.Suspensions {
					t.Logf("  sus[%d]  blocked_by=%d wait=%v", j, s.BlockedBy, s.Resume.Sub(s.Suspend))
				}
				break
			}
		}
	}
}
