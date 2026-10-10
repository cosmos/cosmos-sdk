package ante_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const rekeyAnteChainID = "rekey-ante-chain"

// rekeyToMlDsa65 stores a secp256k1 pubkey on the account (as a first tx would)
// and then rotates the account to a fresh ML-DSA-65 key through
// MsgChangePubKey, the same path a real rotation takes.
func rekeyToMlDsa65(t *testing.T, suite *AnteTestSuite, ta TestAccount) cryptotypes.PrivKey {
	t.Helper()

	sk, err := mldsa65.GenPrivKey()
	require.NoError(t, err)
	rekeyTo(t, suite, ta, sk.PubKey(), &sk)
	return &sk
}

// rekeyToMlDsa65Multisig rotates the account to a threshold-of-n multisig of
// fresh ML-DSA-65 keys and returns the multisig pubkey and the subkeys.
func rekeyToMlDsa65Multisig(t *testing.T, suite *AnteTestSuite, ta TestAccount, threshold, n int) (*kmultisig.LegacyAminoPubKey, []cryptotypes.PrivKey) {
	t.Helper()

	sks := make([]cryptotypes.PrivKey, n)
	pks := make([]cryptotypes.PubKey, n)
	for i := range n {
		sk, err := mldsa65.GenPrivKey()
		require.NoError(t, err)
		sks[i], pks[i] = &sk, sk.PubKey()
	}
	msPk := kmultisig.NewLegacyAminoPubKey(threshold, pks)
	rekeyTo(t, suite, ta, msPk, sks[:threshold]...)
	return msPk, sks
}

// rekeyTo stores a secp256k1 pubkey on the account (as a first tx would) and
// then rotates the account to newPk through MsgChangePubKey, the same path a
// real rotation takes. signers sign the proof: one signer gives a
// single-signature proof, a multisig newPk a MultiSignatureData.
func rekeyTo(t *testing.T, suite *AnteTestSuite, ta TestAccount, newPk cryptotypes.PubKey, signers ...cryptotypes.PrivKey) {
	t.Helper()

	ctx := suite.ctx
	acc := suite.accountKeeper.GetAccount(ctx, ta.acc.GetAddress())
	require.NoError(t, acc.SetPubKey(ta.priv.PubKey()))
	suite.accountKeeper.SetAccount(ctx, acc)

	params := suite.accountKeeper.GetParams(ctx)
	params.PubKeyChangeEnabled = true
	require.NoError(t, suite.accountKeeper.Params.Set(ctx, params))

	anyPk, err := codectypes.NewAnyWithValue(newPk)
	require.NoError(t, err)
	doc := authtypes.ChangePubKeyProofDoc{
		ChainId:       ctx.ChainID(),
		AccountNumber: acc.GetAccountNumber(),
		Address:       acc.GetAddress().String(),
		NewPubKey:     anyPk,
	}
	signBytes, err := authtypes.ChangePubKeyProofSignBytes(suite.accountKeeper.AddressCodec(), doc, newPk)
	require.NoError(t, err)

	sign := func(sk cryptotypes.PrivKey) *signing.SingleSignatureData {
		sig, err := sk.Sign(signBytes)
		require.NoError(t, err)
		return &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, Signature: sig}
	}
	var proof signing.SignatureData
	if msPk, ok := newPk.(*kmultisig.LegacyAminoPubKey); ok {
		keys := msPk.GetPubKeys()
		multi := multisig.NewMultisig(len(keys))
		for _, sk := range signers {
			require.NoError(t, multisig.AddSignatureV2(multi, signing.SignatureV2{PubKey: sk.PubKey(), Data: sign(sk)}, keys))
		}
		proof = multi
	} else {
		require.Len(t, signers, 1)
		proof = sign(signers[0])
	}
	proofBz, err := signing.SignatureDataToProto(proof).Marshal()
	require.NoError(t, err)

	_, err = keeper.NewMsgServerImpl(suite.accountKeeper).ChangePubKey(ctx, &authtypes.MsgChangePubKey{
		Address:   acc.GetAddress().String(),
		NewPubKey: anyPk,
		Proof:     proofBz,
	})
	require.NoError(t, err)

	got := suite.accountKeeper.GetAccount(ctx, acc.GetAddress())
	require.True(t, got.GetPubKey().Equals(newPk))
	require.NotEqual(t, sdk.AccAddress(newPk.Address()), got.GetAddress())
}

// signTxFor builds and signs a single-signer tx for signer with priv. The
// signer address is passed explicitly because a rekeyed account's address does
// not derive from priv. When omitPubKey is set, SignerInfo carries no pubkey in
// both signing rounds, so the signed auth info matches the final tx.
func signTxFor(
	t *testing.T, suite *AnteTestSuite, signer sdk.AccAddress, priv cryptotypes.PrivKey,
	accNum, seq uint64, omitPubKey, unordered bool,
) xauthsigning.Tx {
	t.Helper()
	return signTxForWithGas(t, suite, signer, priv, accNum, seq, omitPubKey, unordered, testdata.NewTestGasLimit())
}

// signTxForWithGas is signTxFor with an explicit gas limit.
func signTxForWithGas(
	t *testing.T, suite *AnteTestSuite, signer sdk.AccAddress, priv cryptotypes.PrivKey,
	accNum, seq uint64, omitPubKey, unordered bool, gasLimit uint64,
) xauthsigning.Tx {
	t.Helper()

	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
	require.NoError(t, suite.txBuilder.SetMsgs(testdata.NewTestMsg(signer)))
	suite.txBuilder.SetFeeAmount(testdata.NewTestFeeAmount())
	suite.txBuilder.SetGasLimit(gasLimit)
	if unordered {
		suite.txBuilder.SetUnordered(true)
		suite.txBuilder.SetTimeoutTimestamp(suite.ctx.BlockTime().Add(time.Minute))
	}

	var pk cryptotypes.PubKey
	if !omitPubKey {
		pk = priv.PubKey()
	}
	require.NoError(t, suite.txBuilder.SetSignatures(signing.SignatureV2{
		PubKey:   pk,
		Data:     &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT},
		Sequence: seq,
	}))

	sigV2, err := tx.SignWithPrivKey(
		suite.ctx, signing.SignMode_SIGN_MODE_DIRECT,
		xauthsigning.SignerData{
			Address:       signer.String(),
			ChainID:       suite.ctx.ChainID(),
			AccountNumber: accNum,
			Sequence:      seq,
			PubKey:        priv.PubKey(),
		},
		suite.txBuilder, priv, suite.clientCtx.TxConfig, seq)
	require.NoError(t, err)
	sigV2.PubKey = pk
	require.NoError(t, suite.txBuilder.SetSignatures(sigV2))

	return suite.txBuilder.GetTx()
}

func runAnte(t *testing.T, suite *AnteTestSuite, ctx sdk.Context, sdkTx sdk.Tx, simulate bool) (sdk.Context, error) {
	t.Helper()
	txBytes, err := suite.clientCtx.TxConfig.TxEncoder()(sdkTx)
	require.NoError(t, err)
	return suite.anteHandler(ctx.WithTxBytes(txBytes), sdkTx, simulate)
}

func setupRekeySuite(t *testing.T, unordered bool) (*AnteTestSuite, TestAccount) {
	t.Helper()
	suite := SetupTestSuiteWithUnordered(t, false, unordered)
	suite.ctx = suite.ctx.WithChainID(rekeyAnteChainID).WithBlockTime(time.Unix(1_700_000_000, 0))
	suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	accs := suite.CreateTestAccounts(1)
	return suite, accs[0]
}

func TestSetPubKey_RekeyedAccountSigns(t *testing.T) {
	suite, ta := setupRekeySuite(t, false)
	addr := ta.acc.GetAddress()
	mlSk := rekeyToMlDsa65(t, suite, ta)

	// Signed by the ML-DSA-65 key with that pubkey in SignerInfo.
	acc := suite.accountKeeper.GetAccount(suite.ctx, addr)
	sdkTx := signTxFor(t, suite, addr, mlSk, acc.GetAccountNumber(), acc.GetSequence(), false, false)
	ctx, err := runAnte(t, suite, suite.ctx, sdkTx, false)
	require.NoError(t, err)

	// Same, with the SignerInfo pubkey omitted.
	acc = suite.accountKeeper.GetAccount(ctx, addr)
	require.Equal(t, uint64(1), acc.GetSequence())
	sdkTx = signTxFor(t, suite, addr, mlSk, acc.GetAccountNumber(), acc.GetSequence(), true, false)
	ctx, err = runAnte(t, suite, ctx, sdkTx, false)
	require.NoError(t, err)

	got := suite.accountKeeper.GetAccount(ctx, addr)
	require.Equal(t, uint64(2), got.GetSequence())
	require.True(t, got.GetPubKey().Equals(mlSk.PubKey()))
}

func TestSetPubKey_OldKeyRejectedAfterRotation(t *testing.T) {
	t.Run("ordered", func(t *testing.T) {
		suite, ta := setupRekeySuite(t, false)
		addr := ta.acc.GetAddress()
		mlSk := rekeyToMlDsa65(t, suite, ta)

		acc := suite.accountKeeper.GetAccount(suite.ctx, addr)
		sdkTx := signTxFor(t, suite, addr, ta.priv, acc.GetAccountNumber(), acc.GetSequence(), false, false)
		_, err := runAnte(t, suite, suite.ctx, sdkTx, false)
		require.ErrorIs(t, err, sdkerrors.ErrInvalidPubKey)

		got := suite.accountKeeper.GetAccount(suite.ctx, addr)
		require.True(t, got.GetPubKey().Equals(mlSk.PubKey()))
		require.Equal(t, acc.GetSequence(), got.GetSequence())
	})

	t.Run("unordered", func(t *testing.T) {
		suite, ta := setupRekeySuite(t, true)
		addr := ta.acc.GetAddress()
		mlSk := rekeyToMlDsa65(t, suite, ta)

		acc := suite.accountKeeper.GetAccount(suite.ctx, addr)
		sdkTx := signTxFor(t, suite, addr, ta.priv, acc.GetAccountNumber(), 0, false, true)
		_, err := runAnte(t, suite, suite.ctx, sdkTx, false)
		require.ErrorIs(t, err, sdkerrors.ErrInvalidPubKey)

		got := suite.accountKeeper.GetAccount(suite.ctx, addr)
		require.True(t, got.GetPubKey().Equals(mlSk.PubKey()))
	})

	// With the pubkey omitted, SetPubKeyDecorator has nothing to compare and
	// skips the signer; signature verification against the stored key
	// rejects it.
	for _, unordered := range []bool{false, true} {
		name := "ordered, pubkey omitted"
		if unordered {
			name = "unordered, pubkey omitted"
		}
		t.Run(name, func(t *testing.T) {
			suite, ta := setupRekeySuite(t, unordered)
			addr := ta.acc.GetAddress()
			mlSk := rekeyToMlDsa65(t, suite, ta)

			acc := suite.accountKeeper.GetAccount(suite.ctx, addr)
			seq := acc.GetSequence()
			if unordered {
				seq = 0
			}
			sdkTx := signTxFor(t, suite, addr, ta.priv, acc.GetAccountNumber(), seq, true, unordered)
			_, err := runAnte(t, suite, suite.ctx, sdkTx, false)
			require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)

			got := suite.accountKeeper.GetAccount(suite.ctx, addr)
			require.True(t, got.GetPubKey().Equals(mlSk.PubKey()))
			require.Equal(t, acc.GetSequence(), got.GetSequence())
		})
	}
}

func TestSetPubKey_UnrekeyedUnchanged(t *testing.T) {
	t.Run("existing account, mismatched tx pubkey", func(t *testing.T) {
		suite, ta := setupRekeySuite(t, false)
		addr := ta.acc.GetAddress()

		// A first tx stores the account's natural pubkey.
		sdkTx := signTxFor(t, suite, addr, ta.priv, ta.acc.GetAccountNumber(), 0, false, false)
		ctx, err := runAnte(t, suite, suite.ctx, sdkTx, false)
		require.NoError(t, err)
		require.True(t, suite.accountKeeper.GetAccount(ctx, addr).GetPubKey().Equals(ta.priv.PubKey()))

		other := secp256k1.GenPrivKey()
		sdkTx = signTxFor(t, suite, addr, other, ta.acc.GetAccountNumber(), 1, false, false)
		_, err = runAnte(t, suite, ctx, sdkTx, false)
		require.ErrorIs(t, err, sdkerrors.ErrInvalidPubKey)
	})

	t.Run("new account, tx pubkey does not hash to address", func(t *testing.T) {
		suite, ta := setupRekeySuite(t, false)
		addr := ta.acc.GetAddress()
		require.Nil(t, suite.accountKeeper.GetAccount(suite.ctx, addr).GetPubKey())

		other := secp256k1.GenPrivKey()
		sdkTx := signTxFor(t, suite, addr, other, ta.acc.GetAccountNumber(), 0, false, false)
		_, err := runAnte(t, suite, suite.ctx, sdkTx, false)
		require.ErrorIs(t, err, sdkerrors.ErrInvalidPubKey)
		require.ErrorContains(t, err, "pubKey does not match signer address")
		require.Nil(t, suite.accountKeeper.GetAccount(suite.ctx, addr).GetPubKey())
	})
}

func TestSimulate_RekeyedMlDsa(t *testing.T) {
	suite, ta := setupRekeySuite(t, false)
	addr := ta.acc.GetAddress()
	mlSk := rekeyToMlDsa65(t, suite, ta)
	acc := suite.accountKeeper.GetAccount(suite.ctx, addr)

	sdkTx := signTxFor(t, suite, addr, mlSk, acc.GetAccountNumber(), acc.GetSequence(), false, false)

	// SetUpContextDecorator installs a fresh gas meter for each run. Simulate
	// on a cached context so its sequence increment does not reach the deliver
	// run.
	cacheCtx, _ := suite.ctx.CacheContext()
	simCtx, err := runAnte(t, suite, cacheCtx, sdkTx, true)
	require.NoError(t, err)
	simulatedGas := simCtx.GasMeter().GasConsumed()

	deliverCtx, err := runAnte(t, suite, suite.ctx, sdkTx, false)
	require.NoError(t, err)
	deliveredGas := deliverCtx.GasMeter().GasConsumed()

	require.GreaterOrEqual(t, simulatedGas, deliveredGas)
}

// A delivered tx whose SignerInfo omits the pubkey has nothing to compare
// against the stored key, so SetPubKeyDecorator must not read the account and
// must consume no gas, as before rekeying.
func TestSetPubKey_NilTxPubKeySkipsAccountRead(t *testing.T) {
	suite, ta := setupRekeySuite(t, false)
	addr := ta.acc.GetAddress()
	acc := suite.accountKeeper.GetAccount(suite.ctx, addr)
	require.NoError(t, acc.SetPubKey(ta.priv.PubKey()))
	suite.accountKeeper.SetAccount(suite.ctx, acc)

	sdkTx := signTxFor(t, suite, addr, ta.priv, acc.GetAccountNumber(), acc.GetSequence(), true, false)

	antehandler := sdk.ChainAnteDecorators(ante.NewSetPubKeyDecorator(suite.accountKeeper))
	ctx := suite.ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
	ctx, err := antehandler(ctx, sdkTx, false)
	require.NoError(t, err)
	require.Zero(t, ctx.GasMeter().GasConsumed())
}

// An existing, never-rekeyed account whose tx carries its own stored pubkey
// still passes the full ante chain.
func TestSetPubKey_StoredKeyResubmittedPasses(t *testing.T) {
	suite, ta := setupRekeySuite(t, false)
	addr := ta.acc.GetAddress()

	sdkTx := signTxFor(t, suite, addr, ta.priv, ta.acc.GetAccountNumber(), 0, false, false)
	ctx, err := runAnte(t, suite, suite.ctx, sdkTx, false)
	require.NoError(t, err)
	require.True(t, suite.accountKeeper.GetAccount(ctx, addr).GetPubKey().Equals(ta.priv.PubKey()))

	sdkTx = signTxFor(t, suite, addr, ta.priv, ta.acc.GetAccountNumber(), 1, false, false)
	_, err = runAnte(t, suite, ctx, sdkTx, false)
	require.NoError(t, err)
}

// Clients simulate with the signer's pubkey and an empty signature
// (client/tx Factory.BuildSimTx). The estimate must still cover the gas a
// signed tx from a rekeyed ML-DSA-65 account consumes when delivered.
func TestSimulate_RekeyedMlDsa_Unsigned(t *testing.T) {
	suite, ta := setupRekeySuite(t, false)
	addr := ta.acc.GetAddress()
	mlSk := rekeyToMlDsa65(t, suite, ta)
	acc := suite.accountKeeper.GetAccount(suite.ctx, addr)

	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
	require.NoError(t, suite.txBuilder.SetMsgs(testdata.NewTestMsg(addr)))
	suite.txBuilder.SetFeeAmount(testdata.NewTestFeeAmount())
	suite.txBuilder.SetGasLimit(testdata.NewTestGasLimit())
	require.NoError(t, suite.txBuilder.SetSignatures(signing.SignatureV2{
		PubKey:   mlSk.PubKey(),
		Data:     &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT},
		Sequence: acc.GetSequence(),
	}))
	simTx := suite.txBuilder.GetTx()

	cacheCtx, _ := suite.ctx.CacheContext()
	simCtx, err := runAnte(t, suite, cacheCtx, simTx, true)
	require.NoError(t, err)
	simulatedGas := simCtx.GasMeter().GasConsumed()

	sdkTx := signTxFor(t, suite, addr, mlSk, acc.GetAccountNumber(), acc.GetSequence(), false, false)
	deliverCtx, err := runAnte(t, suite, suite.ctx, sdkTx, false)
	require.NoError(t, err)
	deliveredGas := deliverCtx.GasMeter().GasConsumed()

	require.GreaterOrEqual(t, simulatedGas, deliveredGas)
}

// A native ML-DSA-65 account simulating its first tx has no stored pubkey, so
// the estimate must size the signature from the SignerInfo pubkey and cover
// the gas the signed tx consumes when delivered.
func TestSimulate_NativeMlDsaFirstTx_Unsigned(t *testing.T) {
	suite, _ := setupRekeySuite(t, false)
	sk, err := mldsa65.GenPrivKey()
	require.NoError(t, err)
	mlSk := &sk
	addr := sdk.AccAddress(mlSk.PubKey().Address())
	acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
	suite.accountKeeper.SetAccount(suite.ctx, acc)
	acc = suite.accountKeeper.GetAccount(suite.ctx, addr)
	require.Nil(t, acc.GetPubKey())

	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
	require.NoError(t, suite.txBuilder.SetMsgs(testdata.NewTestMsg(addr)))
	suite.txBuilder.SetFeeAmount(testdata.NewTestFeeAmount())
	// Storing the several-KB ML-DSA-65 pubkey costs more than the default limit.
	gasLimit := 10 * testdata.NewTestGasLimit()
	suite.txBuilder.SetGasLimit(gasLimit)
	require.NoError(t, suite.txBuilder.SetSignatures(signing.SignatureV2{
		PubKey:   mlSk.PubKey(),
		Data:     &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT},
		Sequence: acc.GetSequence(),
	}))
	simTx := suite.txBuilder.GetTx()

	cacheCtx, _ := suite.ctx.CacheContext()
	simCtx, err := runAnte(t, suite, cacheCtx, simTx, true)
	require.NoError(t, err)
	simulatedGas := simCtx.GasMeter().GasConsumed()

	sdkTx := signTxForWithGas(t, suite, addr, mlSk, acc.GetAccountNumber(), acc.GetSequence(), false, false, gasLimit)
	deliverCtx, err := runAnte(t, suite, suite.ctx, sdkTx, false)
	require.NoError(t, err)
	deliveredGas := deliverCtx.GasMeter().GasConsumed()

	require.GreaterOrEqual(t, simulatedGas, deliveredGas)
}

// Simulating an unsigned tx from an account rekeyed to an ML-DSA-65 multisig
// (multisig pubkey plus empty sub-signatures, as client/tx
// Factory.BuildSimTx produces) must cover the gas the signed tx consumes.
func TestSimulate_RekeyedMlDsaMultisig_Unsigned(t *testing.T) {
	const threshold, n = 2, 3
	suite, ta := setupRekeySuite(t, false)
	addr := ta.acc.GetAddress()
	msPk, sks := rekeyToMlDsa65Multisig(t, suite, ta, threshold, n)
	acc := suite.accountKeeper.GetAccount(suite.ctx, addr)

	newTxBuilder := func() {
		suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
		require.NoError(t, suite.txBuilder.SetMsgs(testdata.NewTestMsg(addr)))
		suite.txBuilder.SetFeeAmount(testdata.NewTestFeeAmount())
		// ML-DSA-65 signatures and pubkeys are several KB each.
		suite.txBuilder.SetGasLimit(10 * testdata.NewTestGasLimit())
	}

	// Unsigned simulation tx.
	newTxBuilder()
	simSigs := cryptotypes.NewCompactBitArray(n)
	simData := &signing.MultiSignatureData{BitArray: simSigs}
	for i := range threshold {
		simSigs.SetIndex(i, true)
		simData.Signatures = append(simData.Signatures,
			&signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON})
	}
	require.NoError(t, suite.txBuilder.SetSignatures(signing.SignatureV2{
		PubKey: msPk, Data: simData, Sequence: acc.GetSequence(),
	}))
	simTx := suite.txBuilder.GetTx()

	cacheCtx, _ := suite.ctx.CacheContext()
	simCtx, err := runAnte(t, suite, cacheCtx, simTx, true)
	require.NoError(t, err)
	simulatedGas := simCtx.GasMeter().GasConsumed()

	// Signed tx: threshold subkeys sign in amino JSON, which multisig needs.
	newTxBuilder()
	require.NoError(t, suite.txBuilder.SetSignatures(signing.SignatureV2{
		PubKey: msPk, Data: multisig.NewMultisig(n), Sequence: acc.GetSequence(),
	}))
	multi := multisig.NewMultisig(n)
	for _, sk := range sks[:threshold] {
		sigV2, err := tx.SignWithPrivKey(
			suite.ctx, signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
			xauthsigning.SignerData{
				Address:       addr.String(),
				ChainID:       suite.ctx.ChainID(),
				AccountNumber: acc.GetAccountNumber(),
				Sequence:      acc.GetSequence(),
				PubKey:        msPk,
			},
			suite.txBuilder, sk, suite.clientCtx.TxConfig, acc.GetSequence())
		require.NoError(t, err)
		require.NoError(t, multisig.AddSignatureV2(multi, sigV2, msPk.GetPubKeys()))
	}
	require.NoError(t, suite.txBuilder.SetSignatures(signing.SignatureV2{
		PubKey: msPk, Data: multi, Sequence: acc.GetSequence(),
	}))
	sdkTx := suite.txBuilder.GetTx()

	deliverCtx, err := runAnte(t, suite, suite.ctx, sdkTx, false)
	require.NoError(t, err)
	deliveredGas := deliverCtx.GasMeter().GasConsumed()

	require.GreaterOrEqual(t, simulatedGas, deliveredGas)
}

// Sizing the simulated signature from the stored pubkey must not change gas
// for a plain secp256k1 account: delivered tx-size gas is exactly
// TxSizeCostPerByte*len(txBytes), and the simulate placeholder is still a
// 64-byte signature with the stored pubkey.
func TestConsumeTxSizeGas_Secp256k1Unchanged(t *testing.T) {
	suite, ta := setupRekeySuite(t, false)
	addr := ta.acc.GetAddress()
	acc := suite.accountKeeper.GetAccount(suite.ctx, addr)
	require.NoError(t, acc.SetPubKey(ta.priv.PubKey()))
	suite.accountKeeper.SetAccount(suite.ctx, acc)
	params := suite.accountKeeper.GetParams(suite.ctx)

	// gasOf returns the gas f consumes on a fresh meter.
	gasOf := func(f func(ctx sdk.Context)) storetypes.Gas {
		ctx := suite.ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
		f(ctx)
		return ctx.GasMeter().GasConsumed()
	}
	paramsGas := gasOf(func(ctx sdk.Context) { suite.accountKeeper.GetParams(ctx) })
	accGas := gasOf(func(ctx sdk.Context) { suite.accountKeeper.GetAccount(ctx, addr) })

	decorator := sdk.ChainAnteDecorators(ante.NewConsumeGasForTxSizeDecorator(suite.accountKeeper))
	run := func(sdkTx sdk.Tx, simulate bool) (storetypes.Gas, int) {
		txBytes, err := suite.clientCtx.TxConfig.TxEncoder()(sdkTx)
		require.NoError(t, err)
		consumed := gasOf(func(ctx sdk.Context) {
			_, err = decorator(ctx.WithTxBytes(txBytes), sdkTx, simulate)
		})
		require.NoError(t, err)
		return consumed, len(txBytes)
	}

	signed := signTxFor(t, suite, addr, ta.priv, acc.GetAccountNumber(), acc.GetSequence(), false, false)
	gas, size := run(signed, false)
	require.Equal(t, paramsGas+params.TxSizeCostPerByte*storetypes.Gas(size), gas)

	txBuilder, err := suite.clientCtx.TxConfig.WrapTxBuilder(signed)
	require.NoError(t, err)
	require.NoError(t, txBuilder.SetSignatures(signing.SignatureV2{
		PubKey:   ta.priv.PubKey(),
		Data:     &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT},
		Sequence: acc.GetSequence(),
	}))
	simGas, simSize := run(txBuilder.GetTx(), true)
	placeholder := legacy.Cdc.MustMarshal(legacytx.StdSignature{ //nolint:staticcheck // SA1019: legacytx.StdSignature is deprecated
		Signature: make([]byte, 64),
		PubKey:    ta.priv.PubKey(),
	})
	require.Equal(t,
		paramsGas+accGas+params.TxSizeCostPerByte*storetypes.Gas(simSize+len(placeholder)+6),
		simGas)
}
