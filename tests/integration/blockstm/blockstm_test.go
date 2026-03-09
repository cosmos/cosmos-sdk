package blockstm_test

import (
	"bytes"
	"strings"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/baseapp/txnrunner"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
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

// TestBlockSTM_AccountCreationPanics validates no recoverable panics occur during
// new account creation happening in parallel via block-stm.
func TestBlockSTM_AccountCreationPanics(t *testing.T) {
	numSenders := 50

	// Generate sender addresses
	senderAddrs := make([]sdk.AccAddress, numSenders)
	for i := range numSenders {
		senderAddrs[i] = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	}

	// Use a capturing logger to detect panics
	var logBuf bytes.Buffer
	logger := log.NewLogger(&logBuf, log.OutputJSONOption())

	// Create store keys for auth and bank
	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey, banktypes.StoreKey)

	// Create codec with auth and bank interfaces registered
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{})
	cdc := encCfg.Codec
	txConfig := encCfg.TxConfig

	// Create BaseApp
	bApp := baseapp.NewBaseApp("blockstm-test", logger, dbm.NewMemDB(), txConfig.TxDecoder(), baseapp.SetChainID("blockstm-test"))
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
	)

	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		map[string]bool{authority.String(): false},
		log.NewNopLogger(),
	)

	// Set InitChainer with default genesis for auth and bank
	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)

	bApp.SetInitChainer(func(ctx sdk.Context, _ *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
		authModule.InitGenesis(ctx, cdc, authModule.DefaultGenesis(cdc))
		bankModule.InitGenesis(ctx, cdc, bankModule.DefaultGenesis(cdc))
		return &abci.ResponseInitChain{}, nil
	})

	// Register bank MsgServer for FinalizeBlock message routing
	banktypes.RegisterMsgServer(bApp.MsgServiceRouter(), bankkeeper.NewMsgServerImpl(bankKeeper))

	// Initialize the chain
	require.NoError(t, bApp.LoadLatestVersion())
	_, err := bApp.InitChain(&abci.RequestInitChain{ChainId: "blockstm-test"})
	require.NoError(t, err)

	// FinalizeBlock (without Commit) keeps finalizeBlockState alive for direct keeper writes
	_, err = bApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: bApp.LastBlockHeight() + 1})
	require.NoError(t, err)

	// Fund all sender accounts with foocoin
	ctx := bApp.NewContext(false)
	for _, addr := range senderAddrs {
		require.NoError(t, testutil.FundAccount(ctx, bankKeeper, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 1000))))
	}

	// Persist funded accounts
	_, err = bApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: bApp.LastBlockHeight() + 1})
	require.NoError(t, err)
	_, err = bApp.Commit()
	require.NoError(t, err)

	// Enable BlockSTM runner with high parallelism
	storeKeys := make([]storetypes.StoreKey, 0, len(keys))
	for _, key := range keys {
		storeKeys = append(storeKeys, key)
	}
	runner := txnrunner.NewSTMRunner(
		txConfig.TxDecoder(),
		storeKeys,
		8,     // high worker count to maximize parallelism
		false, // no pre-estimation to avoid serialization hints
		func(_ storetypes.MultiStore) string { return sdk.DefaultBondDenom },
	)
	bApp.SetBlockSTMTxRunner(runner)

	// Generate unique destination addresses (new accounts, not in genesis)
	recipientAddrs := make([]sdk.AccAddress, numSenders)
	for i := range numSenders {
		recipientAddrs[i] = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	}

	// Build unsigned transactions — no ante handler is configured, so signatures are not needed.
	txBytes := make([][]byte, numSenders)
	for i := range numSenders {
		msg := banktypes.NewMsgSend(senderAddrs[i], recipientAddrs[i], sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)))
		txBuilder := txConfig.NewTxBuilder()
		require.NoError(t, txBuilder.SetMsgs(msg))
		txBytes[i], err = txConfig.TxEncoder()(txBuilder.GetTx())
		require.NoError(t, err)
	}

	// Clear the log buffer before executing the block with BlockSTM
	logBuf.Reset()

	// Execute the block with BlockSTM - all transactions create new accounts in parallel.
	blockRes, err := bApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: bApp.LastBlockHeight() + 1,
		Txs:    txBytes,
	})
	require.NoError(t, err)
	require.Len(t, blockRes.TxResults, numSenders)

	// All transactions should succeed (BlockSTM re-executes on conflict).
	for i, result := range blockRes.TxResults {
		require.Equal(t, uint32(0), result.Code,
			"tx %d should succeed, got code=%d log=%s", i, result.Code, result.Log)
	}

	// Check the log output for evidence of panics from the uniqueness constraint violation.
	logOutput := logBuf.String()
	panicCount := strings.Count(logOutput, "panic recovered in runTx")
	uniquenessViolationCount := strings.Count(logOutput, "uniqueness constraint violation")

	t.Logf("Log output panics: %d recovered panics, %d uniqueness violations", panicCount, uniquenessViolationCount)
	if panicCount > 0 {
		lines := strings.Split(logOutput, "\n")
		shown := 0
		for _, line := range lines {
			if strings.Contains(line, "panic recovered") && shown < 3 {
				t.Logf("  panic log: %s", line)
				shown++
			}
		}
	}

	require.Equal(t, panicCount, 0,
		"expected no panic recovery log entries from uniqueness constraint violations during BlockSTM parallel"+
			" execution of account-creating transactions")
}
