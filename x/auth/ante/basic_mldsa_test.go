package ante_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// TestSimulateGasCoversSignatureSize simulates a tx the way client/tx.Factory
// builds one for "--gas auto" (signer pubkey already set, no signature yet),
// then requires the signed tx to succeed with the simulated gas as its gas
// limit, for each key type with and without the pubkey on chain, and for the
// first tx of multisig accounts and of a tx with two signers.
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
	t.Run("first tx", testSimulateGasFirstTx)
}

// testSimulateGasFirstTx covers the first tx of accounts with no pubkey on
// chain, for which simulation takes the pubkey from the tx: multisig accounts,
// and a tx with two signers.
func testSimulateGasFirstTx(t *testing.T) {
	t.Helper()
	genSecp := func(t *testing.T) cryptotypes.PrivKey { t.Helper(); return secp256k1.GenPrivKey() }
	genMLDSA := func(t *testing.T) cryptotypes.PrivKey {
		t.Helper()
		k, err := mldsa65.GenPrivKey()
		require.NoError(t, err)
		return &k
	}

	t.Run("2-of-3 secp256k1 multisig", func(t *testing.T) {
		testMultisigSimulateGas(t, []cryptotypes.PrivKey{genSecp(t), genSecp(t), genSecp(t)})
	})
	t.Run("2-of-3 ml_dsa_65 multisig", func(t *testing.T) {
		testMultisigSimulateGas(t, []cryptotypes.PrivKey{genMLDSA(t), genMLDSA(t), genMLDSA(t)})
	})
	t.Run("secp256k1 and ml_dsa_65 signers", func(t *testing.T) {
		suite := SetupTestSuite(t, false)
		privs := []cryptotypes.PrivKey{genSecp(t), genMLDSA(t)}
		addrs := make([]sdk.AccAddress, len(privs))
		accNums := []uint64{1000, 1001}
		for i, priv := range privs {
			addrs[i] = sdk.AccAddress(priv.PubKey().Address())
			acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addrs[i])
			require.NoError(t, acc.SetAccountNumber(accNums[i]))
			suite.accountKeeper.SetAccount(suite.ctx, acc)
		}
		fee := testdata.NewTestFeeAmount()
		suite.bankKeeper.EXPECT().
			SendCoinsFromAccountToModule(gomock.Any(), addrs[0], authtypes.FeeCollectorName, fee).
			Return(nil).AnyTimes()
		msg := testdata.NewTestMsg(addrs...)

		suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
		require.NoError(t, suite.txBuilder.SetMsgs(msg))
		suite.txBuilder.SetFeeAmount(fee)
		simSigs := make([]signing.SignatureV2, len(privs))
		for i, priv := range privs {
			simSigs[i] = signing.SignatureV2{
				PubKey: priv.PubKey(),
				Data:   &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT},
			}
		}
		require.NoError(t, suite.txBuilder.SetSignatures(simSigs...))
		simulatedGas := simulateGas(t, suite, suite.txBuilder.GetTx())

		suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
		require.NoError(t, suite.txBuilder.SetMsgs(msg))
		suite.txBuilder.SetFeeAmount(fee)
		suite.txBuilder.SetGasLimit(simulatedGas)
		tx, err := suite.CreateTestTx(suite.ctx, privs, accNums, []uint64{0, 0}, suite.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
		require.NoError(t, err)
		deliverWithGas(t, suite, tx, simulatedGas)
	})
}

// testMultisigSimulateGas simulates the first tx of a 2-of-len(privs) multisig
// account the way client/tx.Factory.BuildSimTx does, then delivers the tx
// signed by the first two keys with the simulated gas as its gas limit.
func testMultisigSimulateGas(t *testing.T, privs []cryptotypes.PrivKey) {
	t.Helper()
	suite := SetupTestSuite(t, false)

	pubKeys := make([]cryptotypes.PubKey, len(privs))
	for i, priv := range privs {
		pubKeys[i] = priv.PubKey()
	}
	const threshold = 2
	multiKey := kmultisig.NewLegacyAminoPubKey(threshold, pubKeys)
	addr := sdk.AccAddress(multiKey.Address())
	const accNum = uint64(1000)
	acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
	require.NoError(t, acc.SetAccountNumber(accNum))
	suite.accountKeeper.SetAccount(suite.ctx, acc)

	fee := testdata.NewTestFeeAmount()
	suite.bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), addr, authtypes.FeeCollectorName, fee).
		Return(nil).AnyTimes()
	msg := testdata.NewTestMsg(addr)

	// The signer info of a multisig signed by the first threshold keys, with
	// empty sub-signatures, as BuildSimTx sets it.
	emptyMultiSig := func() signing.SignatureV2 {
		bitArray := cryptotypes.NewCompactBitArray(len(pubKeys))
		subSigs := make([]signing.SignatureData, threshold)
		for i := range threshold {
			bitArray.SetIndex(i, true)
			subSigs[i] = &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT}
		}
		return signing.SignatureV2{
			PubKey: multiKey,
			Data:   &signing.MultiSignatureData{BitArray: bitArray, Signatures: subSigs},
		}
	}

	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
	require.NoError(t, suite.txBuilder.SetMsgs(msg))
	suite.txBuilder.SetFeeAmount(fee)
	require.NoError(t, suite.txBuilder.SetSignatures(emptyMultiSig()))
	simulatedGas := simulateGas(t, suite, suite.txBuilder.GetTx())

	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
	require.NoError(t, suite.txBuilder.SetMsgs(msg))
	suite.txBuilder.SetFeeAmount(fee)
	suite.txBuilder.SetGasLimit(simulatedGas)
	require.NoError(t, suite.txBuilder.SetSignatures(emptyMultiSig()))
	multiSigData := multisig.NewMultisig(len(pubKeys))
	signerData := xauthsigning.SignerData{
		Address:       addr.String(),
		ChainID:       suite.ctx.ChainID(),
		AccountNumber: accNum,
		Sequence:      0,
		PubKey:        multiKey,
	}
	for _, priv := range privs[:threshold] {
		sig, err := clienttx.SignWithPrivKey(suite.ctx, signing.SignMode_SIGN_MODE_DIRECT, signerData,
			suite.txBuilder, priv, suite.clientCtx.TxConfig, 0)
		require.NoError(t, err)
		require.NoError(t, multisig.AddSignatureV2(multiSigData, sig, pubKeys))
	}
	require.NoError(t, suite.txBuilder.SetSignatures(signing.SignatureV2{PubKey: multiKey, Data: multiSigData}))
	deliverWithGas(t, suite, suite.txBuilder.GetTx(), simulatedGas)
}

// simulateGas runs the ante handler in simulate mode and returns the gas used.
func simulateGas(t *testing.T, suite *AnteTestSuite, simTx sdk.Tx) uint64 {
	t.Helper()
	simBz, err := suite.clientCtx.TxConfig.TxEncoder()(simTx)
	require.NoError(t, err)
	simCtx, _ := suite.ctx.CacheContext()
	simCtx, err = suite.anteHandler(simCtx.WithTxBytes(simBz), simTx, true)
	require.NoError(t, err)
	return simCtx.GasMeter().GasConsumed()
}

// deliverWithGas runs the ante handler on the signed tx, whose gas limit is
// the simulated gas, and requires it to succeed.
func deliverWithGas(t *testing.T, suite *AnteTestSuite, tx sdk.Tx, simulatedGas uint64) {
	t.Helper()
	txBz, err := suite.clientCtx.TxConfig.TxEncoder()(tx)
	require.NoError(t, err)
	deliverCtx, _ := suite.ctx.CacheContext()
	_, err = suite.anteHandler(deliverCtx.WithTxBytes(txBz), tx, false)
	require.NoError(t, err, "simulated gas %d is not enough for the signed tx", simulatedGas)
}
