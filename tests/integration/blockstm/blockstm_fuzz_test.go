package blockstm_test

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/baseapp/txnrunner"
	"github.com/cosmos/cosmos-sdk/client"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
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
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

const (
	fuzzDenom   = "foocoin"
	fuzzChainID = "blockstm-fuzz"
)

type fuzzAppHarness struct {
	app           *baseapp.BaseApp
	txConfig      client.TxConfig
	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.BaseKeeper
	keys          map[string]*storetypes.KVStoreKey
}

type txStreamConfig struct {
	numBlocks           int
	txsPerBlock         int
	insufficientPercent int
	maxValidAmount      int64
	minInvalidAmount    int64
}

func FuzzBlockSTMAppHashDeterminism(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5, 6, 7, 8})
	f.Add([]byte{255, 1, 200, 4, 1, 30, 60, 9, 11, 13, 15})
	f.Add([]byte("high-contention-seed"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			t.Skip()
		}

		b := fuzzByteReader{data: data}
		seed := int64(binary.LittleEndian.Uint64(b.take(8)))
		r := rand.New(rand.NewSource(seed))

		numFundedAccounts := 2 + int(b.next()%23) // [2,24]
		numBlocks := 1 + int(b.next()%8)          // [1,8]
		txsPerBlock := 1 + int(b.next()%32)       // [1,32]
		workers := 1 + int(b.next()%8)             // [1,8]
		estimate := b.next()%2 == 0
		insufficientPercent := int(b.next()%10)  // [0,9]
		extraRecipientCount := int(b.next() % 6)   // [0,5]

		funded := deterministicAddrs(data, 0xAA, numFundedAccounts)
		recipients := deterministicAddrs(data, 0xBB, numFundedAccounts+extraRecipientCount)
		queryAddrs := uniqueAddrs(append(append([]sdk.AccAddress{}, funded...), recipients...))

		encoderHarness := newFuzzAppHarness(t)
		blocks := buildTxStream(t, encoderHarness.txConfig, r, funded, recipients, txStreamConfig{
			numBlocks:           numBlocks,
			txsPerBlock:         txsPerBlock,
			insufficientPercent: insufficientPercent,
			maxValidAmount:      200,
			minInvalidAmount:    20_000,
		})

		runDifferentialBlockStream(t, blocks, funded, queryAddrs, workers, estimate, 10_000)
	})
}

func TestBlockSTMCrossWorkerInvariance(t *testing.T) {
	seed := int64(73_991_003)
	r := rand.New(rand.NewSource(seed))

	const (
		numFundedAccounts  = 24
		extraRecipients    = 8
		numBlocks          = 5
		txsPerBlock        = 80
		insufficientPercent = 6
	)

	seedBytes := []byte(fmt.Sprintf("cross-worker-seed-%d", seed))
	funded := deterministicAddrs(seedBytes, 0xA1, numFundedAccounts)
	recipients := deterministicAddrs(seedBytes, 0xB2, numFundedAccounts+extraRecipients)
	queryAddrs := uniqueAddrs(append(append([]sdk.AccAddress{}, funded...), recipients...))

	encoderHarness := newFuzzAppHarness(t)
	blocks := buildTxStream(t, encoderHarness.txConfig, r, funded, recipients, txStreamConfig{
		numBlocks:           numBlocks,
		txsPerBlock:         txsPerBlock,
		insufficientPercent: insufficientPercent,
		maxValidAmount:      300,
		minInvalidAmount:    35_000,
	})

	for _, workers := range []int{1, 2, 4, 8} {
		for _, estimate := range []bool{false, true} {
			name := fmt.Sprintf("workers=%d/estimate=%t", workers, estimate)
			t.Run(name, func(t *testing.T) {
				runDifferentialBlockStream(t, blocks, funded, queryAddrs, workers, estimate, 20_000)
			})
		}
	}
}

func runDifferentialBlockStream(
	t *testing.T,
	blocks [][][]byte,
	funded []sdk.AccAddress,
	queryAddrs []sdk.AccAddress,
	workers int,
	estimate bool,
	fundingAmount int64,
) {
	t.Helper()

	seq := newFuzzAppHarness(t)
	stm := newFuzzAppHarness(t)

	fundAccounts(t, seq, funded, fundingAmount)
	fundAccounts(t, stm, funded, fundingAmount)

	stmRunner := txnrunner.NewSTMRunner(
		stm.txConfig.TxDecoder(),
		[]storetypes.StoreKey{
			stm.keys[authtypes.StoreKey],
			stm.keys[banktypes.StoreKey],
		},
		workers,
		estimate,
		func(_ storetypes.MultiStore) string { return fuzzDenom },
	)
	stm.app.SetBlockSTMTxRunner(stmRunner)

	for block := range blocks {
		seqRes, err := seq.app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: seq.app.LastBlockHeight() + 1,
			Txs:    blocks[block],
		})
		require.NoError(t, err)

		stmRes, err := stm.app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: stm.app.LastBlockHeight() + 1,
			Txs:    blocks[block],
		})
		require.NoError(t, err)

		require.Equal(t, seqRes.AppHash, stmRes.AppHash, "block %d app hash mismatch", block)
		require.Equal(t, len(seqRes.TxResults), len(stmRes.TxResults), "block %d tx result length mismatch", block)
		for i := range seqRes.TxResults {
			require.Equal(t, seqRes.TxResults[i].Code, stmRes.TxResults[i].Code, "block %d tx %d code mismatch", block, i)
		}

		_, err = seq.app.Commit()
		require.NoError(t, err)
		_, err = stm.app.Commit()
		require.NoError(t, err)

		require.Equal(t, seq.app.LastCommitID().Hash, stm.app.LastCommitID().Hash, "block %d commit hash mismatch", block)
		assertQueryStateEqual(t, block, seq, stm, queryAddrs)
	}
}

func buildTxStream(
	t *testing.T,
	txConfig client.TxConfig,
	r *rand.Rand,
	funded []sdk.AccAddress,
	recipients []sdk.AccAddress,
	cfg txStreamConfig,
) [][][]byte {
	t.Helper()

	blocks := make([][][]byte, cfg.numBlocks)
	for block := 0; block < cfg.numBlocks; block++ {
		txs := make([][]byte, cfg.txsPerBlock)
		for i := 0; i < cfg.txsPerBlock; i++ {
			from := funded[r.Intn(len(funded))]
			to := recipients[r.Intn(len(recipients))]
			amount := int64(1 + r.Intn(int(cfg.maxValidAmount)))
			if r.Intn(100) < cfg.insufficientPercent {
				amount = cfg.minInvalidAmount + int64(r.Intn(10_000))
			}

			txBytes, err := newSendTxBytes(txConfig, from, to, amount)
			require.NoError(t, err)
			// This differential test assumes antehandler-style pre-validation:
			// all incoming txs are expected to be proper sdk.FeeTx values.
			assertFeeTx(t, txConfig, txBytes)
			txs[i] = txBytes
		}
		blocks[block] = txs
	}

	return blocks
}

func assertFeeTx(t *testing.T, txConfig client.TxConfig, txBytes []byte) {
	t.Helper()

	tx, err := txConfig.TxDecoder()(txBytes)
	require.NoError(t, err)
	_, ok := tx.(sdk.FeeTx)
	require.True(t, ok, "generated tx must implement sdk.FeeTx")
}

func newFuzzAppHarness(t *testing.T) fuzzAppHarness {
	t.Helper()

	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey, banktypes.StoreKey)
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{})
	cdc := encCfg.Codec
	txConfig := encCfg.TxConfig

	bApp := baseapp.NewBaseApp(
		"blockstm-fuzz",
		log.NewNopLogger(),
		dbm.NewMemDB(),
		txConfig.TxDecoder(),
		baseapp.SetChainID(fuzzChainID),
	)
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
	_, err := bApp.InitChain(&abci.RequestInitChain{ChainId: fuzzChainID})
	require.NoError(t, err)
	_, err = bApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: bApp.LastBlockHeight() + 1})
	require.NoError(t, err)

	return fuzzAppHarness{
		app:           bApp,
		txConfig:      txConfig,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		keys:          keys,
	}
}

func fundAccounts(t *testing.T, h fuzzAppHarness, addrs []sdk.AccAddress, amount int64) {
	t.Helper()
	ctx := h.app.NewContext(false)
	coins := sdk.NewCoins(sdk.NewInt64Coin(fuzzDenom, amount))
	for _, addr := range addrs {
		require.NoError(t, banktestutil.FundAccount(ctx, h.bankKeeper, addr, coins))
	}

	_, err := h.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: h.app.LastBlockHeight() + 1})
	require.NoError(t, err)
	_, err = h.app.Commit()
	require.NoError(t, err)
}

func newSendTxBytes(txConfig client.TxConfig, from, to sdk.AccAddress, amount int64) ([]byte, error) {
	builder := txConfig.NewTxBuilder()
	msg := banktypes.NewMsgSend(from, to, sdk.NewCoins(sdk.NewInt64Coin(fuzzDenom, amount)))
	if err := builder.SetMsgs(msg); err != nil {
		return nil, err
	}
	return txConfig.TxEncoder()(builder.GetTx())
}

func assertQueryStateEqual(t *testing.T, block int, seq, stm fuzzAppHarness, addrs []sdk.AccAddress) {
	t.Helper()

	seqCtx, err := seq.app.CreateQueryContext(0, false)
	require.NoError(t, err)
	stmCtx, err := stm.app.CreateQueryContext(0, false)
	require.NoError(t, err)

	for _, addr := range addrs {
		seqBalances := seq.bankKeeper.GetAllBalances(seqCtx, addr)
		stmBalances := stm.bankKeeper.GetAllBalances(stmCtx, addr)
		require.True(t, seqBalances.Equal(stmBalances), "block %d balance mismatch for %s", block, addr.String())

		seqAcc := seq.accountKeeper.GetAccount(seqCtx, addr)
		stmAcc := stm.accountKeeper.GetAccount(stmCtx, addr)
		if seqAcc == nil || stmAcc == nil {
			require.Equal(t, seqAcc == nil, stmAcc == nil, "block %d account existence mismatch for %s", block, addr.String())
			continue
		}

		require.Equal(t, seqAcc.GetAddress().String(), stmAcc.GetAddress().String(), "block %d account address mismatch for %s", block, addr.String())
		require.Equal(t, seqAcc.GetAccountNumber(), stmAcc.GetAccountNumber(), "block %d account number mismatch for %s", block, addr.String())
		require.Equal(t, seqAcc.GetSequence(), stmAcc.GetSequence(), "block %d sequence mismatch for %s", block, addr.String())
	}

	seqSupply := seq.bankKeeper.GetSupply(seqCtx, fuzzDenom)
	stmSupply := stm.bankKeeper.GetSupply(stmCtx, fuzzDenom)
	require.True(t, seqSupply.Equal(stmSupply), "block %d supply mismatch for %s", block, fuzzDenom)
}

func uniqueAddrs(addrs []sdk.AccAddress) []sdk.AccAddress {
	seen := make(map[string]struct{}, len(addrs))
	out := make([]sdk.AccAddress, 0, len(addrs))
	for _, addr := range addrs {
		key := addr.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, addr)
	}
	return out
}

func deterministicAddrs(seed []byte, domain byte, count int) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, count)
	for i := 0; i < count; i++ {
		payload := []byte(fmt.Sprintf("%x-%d-%d", seed, domain, i))
		sum := sha256.Sum256(payload)
		addr := make([]byte, 20)
		copy(addr, sum[:20])
		addrs[i] = sdk.AccAddress(addr)
	}
	return addrs
}

type fuzzByteReader struct {
	data []byte
	pos  int
}

func (r *fuzzByteReader) next() byte {
	v := r.data[r.pos%len(r.data)]
	r.pos++
	return v
}

func (r *fuzzByteReader) take(n int) []byte {
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = r.next()
	}
	return out
}
