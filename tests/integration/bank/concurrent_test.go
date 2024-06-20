package bank_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	_ "github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	_ "github.com/cosmos/cosmos-sdk/x/distribution"
	_ "github.com/cosmos/cosmos-sdk/x/gov"
	_ "github.com/cosmos/cosmos-sdk/x/params"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
)

func Test20685(t *testing.T) {
	var (
		senderKey                      = secp256k1.GenPrivKey()
		senderAddr                     = sdk.AccAddress(senderKey.PubKey().Address())
		otherSenderKey                 = secp256k1.GenPrivKey()
		otherSenderAddr                = sdk.AccAddress(otherSenderKey.PubKey().Address())
		receiverAddr    sdk.AccAddress = bytes.Repeat([]byte{0x1}, 32)
	)
	specs := map[string]struct {
		do     func(t *testing.T, app *runtime.App, txBytes [][]byte)
		expSeq []uint64
	}{
		"simulation, no conflicts - sequential": {
			do: func(t *testing.T, app *runtime.App, txsBytes [][]byte) {
				for _, txBytes := range txsBytes {
					_, _, err := app.Simulate(txBytes)
					require.NoError(t, err)
				}
			},
			expSeq: []uint64{0, 0},
		},
		"checkTX : seq": {
			do: func(t *testing.T, app *runtime.App, txsBytes [][]byte) {
				for _, txBytes := range txsBytes {
					_, err := app.CheckTx(&abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New})
					require.NoError(t, err)
				}
			},
			expSeq: []uint64{1, 1},
		},
		"sim + checkTX; no conflicts; sequential": {
			do: func(t *testing.T, app *runtime.App, txsBytes [][]byte) {
				for _, txBytes := range txsBytes {
					_, _, err := app.Simulate(txBytes)
					require.NoError(t, err)
					_, err = app.CheckTx(&abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New})
					require.NoError(t, err)
				}
			},
			expSeq: []uint64{1, 1},
		},
		"checkTX + sim; conflict on account sequence; sequential": {
			do: func(t *testing.T, app *runtime.App, txsBytes [][]byte) {
				for _, txBytes := range txsBytes {
					_, err := app.CheckTx(&abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New})
					require.NoError(t, err)
					_, _, err = app.Simulate(txBytes)
					require.Error(t, err) // account sequence mismatch in ante handler
				}
			},
			expSeq: []uint64{1, 1},
		},
		"sim vs checkTX; no conflicting TX; concurrent": {
			do: func(t *testing.T, app *runtime.App, txsBytes [][]byte) {
				var (
					wgStart, wgDone sync.WaitGroup
					errCount        atomic.Uint32
				)
				const n = 2
				wgStart.Add(n)
				wgDone.Add(n)
				for i := 0; i < n; i++ {
					go func(i int) {
						wgStart.Done()
						wgStart.Wait() // wait for all routines started
						var err error
						if i%2 == 0 {
							_, _, err = app.Simulate(txsBytes[i])
						} else {
							_, err = app.CheckTx(&abci.RequestCheckTx{Tx: txsBytes[i], Type: abci.CheckTxType_New})
						}
						if err != nil {
							errCount.Add(1)
						}
						wgDone.Done()
					}(i)
				}
				wgDone.Wait() // wait for all routines completed
				assert.Equal(t, uint32(0), errCount.Load())
			},
			expSeq: []uint64{0, 1},
		},
		"sim vs checkTX; potential conflict on account sequence; concurrent": {
			do: func(t *testing.T, app *runtime.App, txsBytes [][]byte) {
				var (
					wgStart, wgDone sync.WaitGroup
					errCount        atomic.Uint32
				)
				const n = 2
				wgStart.Add(n)
				wgDone.Add(n)
				for i := 0; i < n; i++ {
					go func(i int) {
						wgStart.Done()
						wgStart.Wait() // wait for all routines started
						var err error
						if i%2 == 0 {
							_, _, err = app.Simulate(txsBytes[0])
						} else {
							_, err = app.CheckTx(&abci.RequestCheckTx{Tx: txsBytes[0], Type: abci.CheckTxType_New})
						}
						if err != nil {
							errCount.Add(1)
						}
						wgDone.Done()
					}(i)
				}
				wgDone.Wait()                               // wait for all routines completed
				assert.Equal(t, uint32(0), errCount.Load()) // when both TX enter the ante handlers at the same time, there is no account seq conflict. Both have a forked store
			},
			expSeq: []uint64{1, 0},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			for i := 0; i < 10; i++ { // run a couple of times to show this is not random behaviour
				t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
					// setup chain, accounts, and 2 example TX for sender and other sender
					s, _, _, _, _, origSeq, txsBytes := setupSUT(t, senderAddr, receiverAddr, senderKey, otherSenderKey)

					// when
					spec.do(t, s.App, txsBytes)
					// then
					ctx := s.App.NewContext(true)
					assert.Equal(t, spec.expSeq[0], s.AccountKeeper.GetAccount(ctx, senderAddr).GetSequence())      // seq not bumped as no TX delivered
					assert.Equal(t, spec.expSeq[1], s.AccountKeeper.GetAccount(ctx, otherSenderAddr).GetSequence()) // seq not bumped as no TX delivered
					assert.Nil(t, s.AccountKeeper.GetAccount(ctx, receiverAddr))

					// and when committed
					nextBlock(t, s.App)

					// then state is reset as there was no deliver TX
					ctx = s.App.NewContext(true)
					assert.Equal(t, origSeq[0], s.AccountKeeper.GetAccount(ctx, senderAddr).GetSequence())      // seq not bumped as no TX delivered
					assert.Equal(t, origSeq[1], s.AccountKeeper.GetAccount(ctx, otherSenderAddr).GetSequence()) // seq not bumped as no TX delivered
					assert.Nil(t, s.AccountKeeper.GetAccount(ctx, receiverAddr))
				})
			}
		})
	}
}

func setupSUT(t *testing.T, _ sdk.AccAddress, receiverAddr sdk.AccAddress, senderKeys ...*secp256k1.PrivKey) (*suite, *baseapp.BaseApp, sdk.Context, error, sdk.AccountI, []uint64, [][]byte) {
	ga := make([]authtypes.GenesisAccount, len(senderKeys))
	for i, senderKey := range senderKeys {
		senderAddr := sdk.AccAddress(senderKey.PubKey().Address())
		ga[i] = &authtypes.BaseAccount{Address: senderAddr.String()}
	}
	s := createTestSuite(t, ga)
	ctx := s.App.NewContext(false)
	// fund
	for _, senderKey := range senderKeys {
		senderAddr := sdk.AccAddress(senderKey.PubKey().Address())
		require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, senderAddr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 100_000_000))))
	}
	nextBlock(t, s.App)

	txsBz := make([][]byte, len(senderKeys))
	seqs := make([]uint64, len(senderKeys))
	ctx = s.App.NewContext(true)
	for i, senderKey := range senderKeys {
		senderAddr := sdk.AccAddress(senderKey.PubKey().Address())
		senderAccount := s.AccountKeeper.GetAccount(ctx, senderAddr)
		require.NotNil(t, senderAccount)
		origAccNum, origSeq := senderAccount.GetAccountNumber(), senderAccount.GetSequence()
		seqs[i] = origSeq

		// encode an example msg
		sendMsg := types.NewMsgSend(senderAddr, receiverAddr, sdk.Coins{sdk.NewInt64Coin("foocoin", 100)})
		txBytes := encodeAndSign(t, s.TxConfig, []sdk.Msg{sendMsg}, "", []uint64{origAccNum}, []uint64{origSeq}, senderKey)
		txsBz[i] = txBytes
	}
	return &s, nil, ctx, nil, nil, seqs, txsBz
}

func createTestSuite(t *testing.T, genesisAccounts []authtypes.GenesisAccount) suite {
	res := suite{}

	var genAccounts []simtestutil.GenesisAccount
	for _, acc := range genesisAccounts {
		genAccounts = append(genAccounts, simtestutil.GenesisAccount{GenesisAccount: acc})
	}

	startupCfg := simtestutil.DefaultStartUpConfig()
	startupCfg.GenesisAccounts = genAccounts

	app, err := simtestutil.SetupWithConfiguration(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.ParamsModule(),
				configurator.AuthModule(),
				configurator.StakingModule(),
				configurator.TxModule(),
				configurator.ConsensusModule(),
				configurator.BankModule(),
				configurator.GovModule(),
				configurator.DistributionModule(),
			),
			depinject.Supply(log.NewNopLogger()),
		),
		startupCfg, &res.BankKeeper, &res.AccountKeeper, &res.TxConfig)

	res.App = app

	require.NoError(t, err)
	return res
}

func nextBlock(t *testing.T, app *runtime.App) {
	t.Helper()
	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: app.LastBlockHeight() + 1})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)
}

func encodeAndSign(
	t *testing.T,
	txConfig client.TxConfig,
	msgs []sdk.Msg,
	chainID string,
	accNums, accSeqs []uint64,
	priv ...cryptotypes.PrivKey,
) []byte {
	t.Helper()
	tx, err := simtestutil.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txConfig,
		msgs,
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
		simtestutil.DefaultGenTxGas,
		chainID,
		accNums,
		accSeqs,
		priv...,
	)
	require.NoError(t, err)
	txBytes, err := txConfig.TxEncoder()(tx)
	require.Nil(t, err)
	return txBytes
}

type suite struct {
	TxConfig      client.TxConfig
	App           *runtime.App
	AccountKeeper authkeeper.AccountKeeper
	BankKeeper    bankkeeper.Keeper
}
