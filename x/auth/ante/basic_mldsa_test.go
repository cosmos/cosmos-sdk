package ante_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// TestSimulateGasCoversSignatureSize simulates a tx the way client/tx.Factory
// builds one for "--gas auto" (signer pubkey already set, no signature yet),
// then requires the signed tx to succeed when broadcast with gas limit equal
// to the simulated gas, for every supported account key type, both before and
// after the account's pubkey is recorded on chain.
//
// An ml_dsa_65 signature is much larger than a secp256k1 one (3,309 vs 64
// bytes), so simulation must charge tx-size gas for the real signature size.
// Before this fix, ConsumeTxSizeGasDecorator always priced the simulated
// signature as 64 bytes regardless of the signer's actual key type, so this
// test fails for ml_dsa_65 with "out of gas" and passes for secp256k1.
func TestSimulateGasCoversSignatureSize(t *testing.T) {
	keys := map[string]func(t *testing.T) cryptotypes.PrivKey{
		"secp256k1": func(t *testing.T) cryptotypes.PrivKey { t.Helper(); return secp256k1.GenPrivKey() },
		"ml_dsa_65": func(t *testing.T) cryptotypes.PrivKey {
			t.Helper()
			k, err := mldsa65.GenPrivKey()
			require.NoError(t, err)
			return &k
		},
	}

	for name, genKey := range keys {
		for _, pubKeyOnChain := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/pubKeyOnChain=%t", name, pubKeyOnChain), func(t *testing.T) {
				suite := SetupTestSuite(t, false)
				priv := genKey(t)
				addr := sdk.AccAddress(priv.PubKey().Address())

				acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
				require.NoError(t, acc.SetAccountNumber(1000))
				if pubKeyOnChain {
					require.NoError(t, acc.SetPubKey(priv.PubKey()))
				}
				suite.accountKeeper.SetAccount(suite.ctx, acc)

				fee := testdata.NewTestFeeAmount()
				suite.bankKeeper.EXPECT().
					SendCoinsFromAccountToModule(gomock.Any(), addr, authtypes.FeeCollectorName, fee).
					Return(nil).AnyTimes()
				msg := testdata.NewTestMsg(addr)

				// Simulation tx: real pubkey already set (client/tx.Factory.BuildSimTx
				// always knows the signer's pubkey), no signature yet.
				suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
				require.NoError(t, suite.txBuilder.SetMsgs(msg))
				suite.txBuilder.SetFeeAmount(fee)
				require.NoError(t, suite.txBuilder.SetSignatures(signing.SignatureV2{
					PubKey: priv.PubKey(),
					Data:   &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT},
				}))
				simTx := suite.txBuilder.GetTx()
				simBz, err := suite.clientCtx.TxConfig.TxEncoder()(simTx)
				require.NoError(t, err)
				simCtx, _ := suite.ctx.CacheContext()
				simCtx, err = suite.anteHandler(simCtx.WithTxBytes(simBz), simTx, true)
				require.NoError(t, err)
				simulatedGas := simCtx.GasMeter().GasConsumed()

				// The actually-signed tx, broadcast with gas limit == simulated gas,
				// must not run out of gas.
				suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
				require.NoError(t, suite.txBuilder.SetMsgs(msg))
				suite.txBuilder.SetFeeAmount(fee)
				suite.txBuilder.SetGasLimit(simulatedGas)
				tx, err := suite.CreateTestTx(suite.ctx, []cryptotypes.PrivKey{priv}, []uint64{1000}, []uint64{0},
					suite.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
				require.NoError(t, err)
				txBz, err := suite.clientCtx.TxConfig.TxEncoder()(tx)
				require.NoError(t, err)
				deliverCtx, _ := suite.ctx.CacheContext()
				_, err = suite.anteHandler(deliverCtx.WithTxBytes(txBz), tx, false)
				require.NoError(t, err, "simulated gas %d is not enough for the signed tx (pubkey on chain: %v)", simulatedGas, pubKeyOnChain)
			})
		}
	}
}
