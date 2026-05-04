package blockstm_test

import (
	"bytes"
	"strings"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
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

type blockSTMTestApp struct {
	app        *baseapp.BaseApp
	bankKeeper bankkeeper.BaseKeeper
	txConfig   client.TxConfig
}

// TestBlockSTM_AccountCreationPanics validates no recoverable panics occur during
// new account creation happening in parallel via block-stm.
func TestBlockSTM_AccountCreationPanics(t *testing.T) {
	numSenders := 50

	// Generate sender addresses
	senderAddrs := generateAddrs(numSenders)

	// Use a capturing logger to detect panics
	var logBuf bytes.Buffer
	logger := log.NewLogger(&logBuf, log.OutputJSONOption())

	blockSTMApp := newBlockSTMTestApp(t, dbm.NewMemDB(), logger, true)
	initChainAndFundAccounts(t, blockSTMApp, senderAddrs)

	// Generate unique destination addresses (new accounts, not in genesis)
	recipientAddrs := generateAddrs(numSenders)
	txBytes := buildSendTxs(t, blockSTMApp.txConfig, senderAddrs, recipientAddrs)

	// Clear the log buffer before executing the block with BlockSTM
	logBuf.Reset()

	// Execute the block with BlockSTM - all transactions create new accounts in parallel.
	blockRes := finalizeNextBlock(t, blockSTMApp.app, txBytes)
	requireSuccessfulTxResults(t, blockRes.TxResults)

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

func TestBlockSTM_DeterministicAppHash(t *testing.T) {
	numSenders := 50
	db := dbm.NewMemDB()

	senderAddrs := generateAddrs(numSenders)
	recipientAddrs := generateAddrs(numSenders)

	sequentialApp := newBlockSTMTestApp(t, db, log.NewNopLogger(), false)
	initChainAndFundAccounts(t, sequentialApp, senderAddrs)

	txBytes := buildSendTxs(t, sequentialApp.txConfig, senderAddrs, recipientAddrs)
	baseVersion := sequentialApp.app.LastCommitID().Version

	sequentialRes := finalizeNextBlock(t, sequentialApp.app, txBytes)
	requireSuccessfulTxResults(t, sequentialRes.TxResults)

	sequentialCommitID := commitBlock(t, sequentialApp.app)

	blockSTMApp := newBlockSTMTestApp(t, db, log.NewNopLogger(), true)
	require.NoError(t, blockSTMApp.app.LoadVersion(baseVersion))

	blockSTMRes := finalizeNextBlock(t, blockSTMApp.app, txBytes)
	requireSuccessfulTxResults(t, blockSTMRes.TxResults)

	blockSTMCommitID := commitBlock(t, blockSTMApp.app)

	require.NotEmpty(t, sequentialRes.AppHash)
	require.NotEmpty(t, sequentialCommitID.Hash)
	require.NotEmpty(t, blockSTMRes.AppHash)
	require.NotEmpty(t, blockSTMCommitID.Hash)

	require.Equal(t, sequentialRes.AppHash, sequentialCommitID.Hash)
	require.Equal(t, sequentialRes.AppHash, blockSTMRes.AppHash)
	require.Equal(t, sequentialCommitID, blockSTMCommitID)
}

func newBlockSTMTestApp(t *testing.T, db dbm.DB, logger log.Logger, enableBlockSTM bool) blockSTMTestApp {
	t.Helper()

	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey, banktypes.StoreKey)
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{})
	cdc := encCfg.Codec

	bApp := baseapp.NewBaseApp("blockstm-test", logger, db, encCfg.TxConfig.TxDecoder(), baseapp.SetChainID("blockstm-test"))
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

	if enableBlockSTM {
		bApp.SetBlockSTMTxRunner(newTestSTMRunner(
			encCfg.TxConfig.TxDecoder(),
			[]storetypes.StoreKey{keys[authtypes.StoreKey], keys[banktypes.StoreKey]},
			8,
		))
	}

	return blockSTMTestApp{
		app:        bApp,
		bankKeeper: bankKeeper,
		txConfig:   encCfg.TxConfig,
	}
}

func initChainAndFundAccounts(t *testing.T, testApp blockSTMTestApp, senderAddrs []sdk.AccAddress) {
	t.Helper()

	require.NoError(t, testApp.app.LoadLatestVersion())

	_, err := testApp.app.InitChain(&abci.RequestInitChain{ChainId: "blockstm-test"})
	require.NoError(t, err)

	_ = finalizeNextBlock(t, testApp.app, nil)

	ctx := testApp.app.NewContext(false)
	for _, addr := range senderAddrs {
		require.NoError(t, testutil.FundAccount(ctx, testApp.bankKeeper, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 1000))))
	}

	_, _ = finalizeAndCommitNextBlock(t, testApp.app, nil)
}

func buildSendTxs(t *testing.T, txConfig client.TxConfig, senderAddrs, recipientAddrs []sdk.AccAddress) [][]byte {
	t.Helper()
	require.Len(t, recipientAddrs, len(senderAddrs))

	txBytes := make([][]byte, len(senderAddrs))
	for i := range senderAddrs {
		msg := banktypes.NewMsgSend(senderAddrs[i], recipientAddrs[i], sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)))
		txBuilder := txConfig.NewTxBuilder()
		require.NoError(t, txBuilder.SetMsgs(msg))

		bz, err := txConfig.TxEncoder()(txBuilder.GetTx())
		require.NoError(t, err)

		txBytes[i] = bz
	}

	return txBytes
}

func generateAddrs(count int) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, count)
	for i := range count {
		addrs[i] = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	}

	return addrs
}

func requireSuccessfulTxResults(t rapid.TB, txResults []*abci.ExecTxResult) {
	t.Helper()

	for i, result := range txResults {
		require.Equal(t, uint32(0), result.Code,
			"tx %d should succeed, got code=%d log=%s", i, result.Code, result.Log)
	}
}
