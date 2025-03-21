package ante_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestUnorderedAnte(t *testing.T) {
	testPK, _, testAddr := testdata.KeyTestPubAddr()
	testCases := map[string]struct {
		addTxs      func() []sdk.Tx
		runTx       func() sdk.Tx
		blockTime   time.Time
		execMode    sdk.ExecMode
		expectedErr string
	}{
		"normal/ordered tx should just skip": {
			runTx: func() sdk.Tx {
				return genTestTx(t, genTxOptions{})
			},
			blockTime: time.Unix(0, 0),
			execMode:  sdk.ExecModeFinalize,
		},
		"happy case - simple pass": {
			runTx: func() sdk.Tx {
				return genTestTx(t, genTxOptions{unordered: true, timestamp: time.Unix(10, 0)})
			},
			blockTime: time.Unix(0, 0),
			execMode:  sdk.ExecModeFinalize,
		},
		"zero time should fail": {
			runTx: func() sdk.Tx {
				return genTestTx(t, genTxOptions{unordered: true})
			},
			blockTime:   time.Unix(0, 0),
			execMode:    sdk.ExecModeFinalize,
			expectedErr: "unordered transaction must have timeout_timestamp set",
		},
		"timeout before current block time should fail": {
			runTx: func() sdk.Tx {
				return genTestTx(t, genTxOptions{unordered: true, timestamp: time.Unix(7, 0)})
			},
			blockTime:   time.Unix(10, 1),
			execMode:    sdk.ExecModeFinalize,
			expectedErr: "unordered transaction has a timeout_timestamp that has already passed",
		},
		"timeout equal to current block time should pass": {
			runTx: func() sdk.Tx {
				return genTestTx(t, genTxOptions{unordered: true, timestamp: time.Unix(10, 0)})
			},
			blockTime: time.Unix(10, 0),
			execMode:  sdk.ExecModeFinalize,
		},
		"timeout after the max duration should fail": {
			runTx: func() sdk.Tx {
				return genTestTx(t, genTxOptions{unordered: true, timestamp: time.Unix(10, 1).Add(ante.DefaultMaxTimoutDuration)})
			},
			blockTime:   time.Unix(10, 0),
			execMode:    sdk.ExecModeFinalize,
			expectedErr: "unordered tx ttl exceeds",
		},
		"fails if manager has duplicate": {
			addTxs: func() []sdk.Tx {
				tx := genTestTx(
					t,
					genTxOptions{unordered: true, timestamp: time.Unix(10, 0), pk: testPK, addr: testAddr},
				)
				return []sdk.Tx{tx}
			},
			runTx: func() sdk.Tx {
				return genTestTx(
					t,
					genTxOptions{unordered: true, timestamp: time.Unix(10, 0), pk: testPK, addr: testAddr},
				)
			},
			blockTime:   time.Unix(5, 0),
			execMode:    sdk.ExecModeFinalize,
			expectedErr: "already used timeout",
		},
		"duplicate doesn't matter if we're in simulate mode": {
			addTxs: func() []sdk.Tx {
				tx := genTestTx(
					t,
					genTxOptions{unordered: true, timestamp: time.Unix(10, 0), pk: testPK, addr: testAddr},
				)
				return []sdk.Tx{tx}
			},
			runTx: func() sdk.Tx {
				return genTestTx(
					t,
					genTxOptions{unordered: true, timestamp: time.Unix(10, 0), pk: testPK, addr: testAddr},
				)
			},
			blockTime: time.Unix(5, 0),
			execMode:  sdk.ExecModeSimulate,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mockStoreKey := storetypes.NewKVStoreKey("test")
			storeService := runtime.NewKVStoreService(mockStoreKey)
			ctx := testutil.DefaultContextWithDB(
				t,
				mockStoreKey,
				storetypes.NewTransientStoreKey("transient_test"),
			).Ctx.WithBlockTime(tc.blockTime).WithExecMode(tc.execMode)
			mgr := keeper.NewAccountKeeper(
				moduletestutil.MakeTestEncodingConfig().Codec,
				storeService,
				types.ProtoBaseAccount,
				nil,
				authcodec.NewBech32Codec("cosmos"),
				"cosmos",
				types.NewModuleAddress("gov").String(),
			)
			chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(mgr))
			var err error
			if tc.addTxs != nil {
				txs := tc.addTxs()
				for _, tx := range txs {
					ctx, err = chain(ctx, tx, false)
					require.NoError(t, err)
				}
			}
			_, err = chain(ctx, tc.runTx(), false)
			if tc.expectedErr != "" {
				require.ErrorContains(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMultiSignerUnorderedTx(t *testing.T) {
	pk1, _, addr1 := testdata.KeyTestPubAddr()
	pk2, _, addr2 := testdata.KeyTestPubAddr()
	pk3, _, addr3 := testdata.KeyTestPubAddr()

	signerAddrs := []sdk.AccAddress{addr1, addr2, addr3}

	mockStoreKey := storetypes.NewKVStoreKey("test")
	storeService := runtime.NewKVStoreService(mockStoreKey)
	ctx := testutil.DefaultContextWithDB(
		t,
		mockStoreKey,
		storetypes.NewTransientStoreKey("transient_test"),
	).Ctx.WithBlockTime(time.Unix(9, 0))
	mgr := keeper.NewAccountKeeper(
		moduletestutil.MakeTestEncodingConfig().Codec,
		storeService,
		types.ProtoBaseAccount,
		nil,
		authcodec.NewBech32Codec("cosmos"),
		"cosmos",
		types.NewModuleAddress("gov").String(),
	)
	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(mgr))

	timeout := time.Unix(10, 0)
	tx := genMultiSignedUnorderedTx(t, signerAddrs, timeout, []cryptotypes.PrivKey{pk1, pk2, pk3})

	newCtx, err := chain(ctx, tx, false)
	require.NoError(t, err)

	for _, addr := range signerAddrs {
		ok, err := mgr.ContainsUnorderedNonce(newCtx, addr.Bytes(), timeout)
		require.NoError(t, err)
		require.True(t, ok)
	}
}

type genTxOptions struct {
	unordered bool
	timestamp time.Time
	pk        cryptotypes.PrivKey
	addr      sdk.AccAddress
}

func genTestTx(t *testing.T, options genTxOptions) sdk.Tx {
	t.Helper()

	s := SetupTestSuite(t, true)
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	priv1 := options.pk
	addr1 := options.addr
	if options.pk == nil || options.addr == nil {
		priv1, _, addr1 = testdata.KeyTestPubAddr()
	}

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	require.NoError(t, s.txBuilder.SetMsgs(msg))

	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)
	s.txBuilder.SetUnordered(options.unordered)
	s.txBuilder.SetTimeoutTimestamp(options.timestamp)

	privKeys, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(s.ctx, privKeys, accNums, accSeqs, s.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
	require.NoError(t, err)

	require.NoError(t, err)

	return tx
}

func genMultiSignedUnorderedTx(t *testing.T, addrs []sdk.AccAddress, ts time.Time, pks []cryptotypes.PrivKey) sdk.Tx {
	t.Helper()

	s := SetupTestSuite(t, true)
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	// msg and signatures
	msgs := make([]sdk.Msg, 0, len(addrs))
	for _, addr := range addrs {
		msgs = append(msgs, testdata.NewTestMsg(addr))
	}
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	require.NoError(t, s.txBuilder.SetMsgs(msgs...))

	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)
	s.txBuilder.SetUnordered(true)
	s.txBuilder.SetTimeoutTimestamp(ts)

	accNums := make([]uint64, len(pks))
	accSeqs := make([]uint64, len(pks))
	tx, err := s.CreateTestTx(s.ctx, pks, accNums, accSeqs, s.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
	require.NoError(t, err)

	require.NoError(t, err)

	return tx
}
