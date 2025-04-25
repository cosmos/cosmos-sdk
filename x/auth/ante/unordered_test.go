package ante_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	txmodule "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestSigVerification_UnorderedTxs(t *testing.T) {
	var suite *AnteTestSuite
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	var priv1, priv2, priv3 cryptotypes.PrivKey
	var antehandler sdk.AnteHandler
	var defaultSignMode signing.SignMode
	var accs []sdk.AccountI
	var msgs []sdk.Msg
	reset := func(isCheckTx, withUnordered bool) {
		suite = SetupTestSuiteWithUnordered(t, isCheckTx, withUnordered)
		suite.txBankKeeper.EXPECT().DenomMetadata(gomock.Any(), gomock.Any()).Return(&banktypes.QueryDenomMetadataResponse{}, nil).AnyTimes()

		enabledSignModes := []signing.SignMode{signing.SignMode_SIGN_MODE_DIRECT, signing.SignMode_SIGN_MODE_TEXTUAL, signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON}
		// Since TEXTUAL is not enabled by default, we create a custom TxConfig
		// here which includes it.
		txConfigOpts := authtx.ConfigOptions{
			TextualCoinMetadataQueryFn: txmodule.NewGRPCCoinMetadataQueryFn(suite.clientCtx),
			EnabledSignModes:           enabledSignModes,
		}
		var err error
		suite.clientCtx.TxConfig, err = authtx.NewTxConfigWithOptions(
			codec.NewProtoCodec(suite.encCfg.InterfaceRegistry),
			txConfigOpts,
		)
		require.NoError(t, err)
		suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

		// make block height non-zero to ensure account numbers part of signBytes
		suite.ctx = suite.ctx.WithBlockHeight(1)

		// keys and addresses
		pk1, _, addr1 := testdata.KeyTestPubAddr()
		pk2, _, addr2 := testdata.KeyTestPubAddr()
		pk3, _, addr3 := testdata.KeyTestPubAddr()
		priv1, priv2, priv3 = pk1, pk2, pk3

		addrs := []sdk.AccAddress{addr1, addr2, addr3}

		msgs = make([]sdk.Msg, len(addrs))
		accs = make([]sdk.AccountI, len(addrs))
		// set accounts and create msg for each address
		for i, addr := range addrs {
			acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
			require.NoError(t, acc.SetAccountNumber(uint64(i)+1000))
			suite.accountKeeper.SetAccount(suite.ctx, acc)
			msgs[i] = testdata.NewTestMsg(addr)
			accs[i] = acc
		}

		spkd := ante.NewSetPubKeyDecorator(suite.accountKeeper)
		txConfigOpts = authtx.ConfigOptions{
			TextualCoinMetadataQueryFn: txmodule.NewBankKeeperCoinMetadataQueryFn(suite.txBankKeeper),
			EnabledSignModes:           enabledSignModes,
		}
		anteTxConfig, err := authtx.NewTxConfigWithOptions(
			codec.NewProtoCodec(suite.encCfg.InterfaceRegistry),
			txConfigOpts,
		)
		require.NoError(t, err)
		svd := ante.NewSigVerificationDecorator(suite.accountKeeper, anteTxConfig.SignModeHandler())
		antehandler = sdk.ChainAnteDecorators(spkd, svd)
		defaultSignMode, err = authsign.APISignModeToInternal(anteTxConfig.SignModeHandler().DefaultMode())
		require.NoError(t, err)
	}

	testCases := map[string]struct {
		unorderedDisabled bool
		unordered         bool
		timeout           time.Time
		blockTime         time.Time
		duplicate         bool
		execMode          sdk.ExecMode
		expectedErr       string
	}{
		"normal/ordered tx should just skip": {
			unordered: false,
			blockTime: time.Unix(0, 0),
			execMode:  sdk.ExecModeFinalize,
		},
		"normal/ordered tx should just skip with unordered disabled too": {
			unorderedDisabled: true,
			unordered:         false,
			blockTime:         time.Unix(0, 0),
			execMode:          sdk.ExecModeFinalize,
		},
		"happy case": {
			unordered: true,
			timeout:   time.Unix(10, 0),
			blockTime: time.Unix(0, 0),
			execMode:  sdk.ExecModeFinalize,
		},
		"zero time should fail": {
			unordered:   true,
			blockTime:   time.Unix(10, 0),
			execMode:    sdk.ExecModeFinalize,
			expectedErr: "unordered transaction must have timeout_timestamp set",
		},
		"fail if tx is unordered but unordered is disabled": {
			unorderedDisabled: true,
			unordered:         true,
			blockTime:         time.Unix(10, 0),
			execMode:          sdk.ExecModeFinalize,
			expectedErr:       "unordered transactions are not enabled",
		},
		"timeout before current block time should fail": {
			unordered:   true,
			timeout:     time.Unix(7, 0),
			blockTime:   time.Unix(10, 1),
			execMode:    sdk.ExecModeFinalize,
			expectedErr: "unordered transaction has a timeout_timestamp that has already passed",
		},
		"timeout equal to current block time should pass": {
			unordered: true,
			timeout:   time.Unix(10, 0),
			blockTime: time.Unix(10, 0),
			execMode:  sdk.ExecModeFinalize,
		},
		"timeout after the max duration should fail": {
			unordered:   true,
			timeout:     time.Unix(10, 1).Add(ante.DefaultMaxTimeoutDuration),
			blockTime:   time.Unix(10, 0),
			execMode:    sdk.ExecModeFinalize,
			expectedErr: "unordered tx ttl exceeds",
		},
		"fails if manager has duplicate": {
			unordered:   true,
			timeout:     time.Unix(10, 0),
			duplicate:   true,
			blockTime:   time.Unix(5, 0),
			execMode:    sdk.ExecModeFinalize,
			expectedErr: "already used timeout",
		},
		"duplicate doesn't matter if we're in simulate mode": {
			unordered: true,
			timeout:   time.Unix(10, 0),
			duplicate: true,
			blockTime: time.Unix(5, 0),
			execMode:  sdk.ExecModeSimulate,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			reset(tc.execMode == sdk.ExecModeCheck, !tc.unorderedDisabled)
			ctx := suite.ctx.WithBlockTime(tc.blockTime).WithExecMode(tc.execMode).WithIsSigverifyTx(true)

			suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder() // Create new txBuilder for each test
			require.NoError(t, suite.txBuilder.SetMsgs(msgs...))
			suite.txBuilder.SetFeeAmount(feeAmount)
			suite.txBuilder.SetGasLimit(gasLimit)
			tx, err := suite.CreateTestUnorderedTx(
				suite.ctx,
				[]cryptotypes.PrivKey{priv1, priv2, priv3},
				[]uint64{accs[0].GetAccountNumber(), accs[1].GetAccountNumber(), accs[2].GetAccountNumber()},
				[]uint64{0, 0, 0},
				suite.ctx.ChainID(),
				defaultSignMode,
				tc.unordered,
				tc.timeout,
			)
			require.NoError(t, err)
			txBytes, err := suite.clientCtx.TxConfig.TxEncoder()(tx)
			require.NoError(t, err)
			ctx = ctx.WithTxBytes(txBytes)

			simulate := tc.execMode == sdk.ExecModeSimulate

			if tc.duplicate {
				_, err = antehandler(ctx, tx, simulate)
				require.NoError(t, err)
			}

			_, err = antehandler(ctx, tx, simulate)
			if tc.expectedErr != "" {
				require.ErrorContains(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
