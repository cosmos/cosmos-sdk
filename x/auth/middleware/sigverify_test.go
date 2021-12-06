package middleware_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (s *MWTestSuite) TestSetPubKey() {
	ctx := s.SetupTest(true) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()
	require := s.Require()
	txHandler := middleware.ComposeMiddlewares(
		noopTxHandler,
		middleware.SetPubKeyMiddleware(s.app.AccountKeeper),
	)

	// keys and addresses
	priv1, pub1, addr1 := testdata.KeyTestPubAddr()
	priv2, pub2, addr2 := testdata.KeyTestPubAddr()
	priv3, pub3, addr3 := testdata.KeyTestPubAddrSecp256R1(require)

	addrs := []sdk.AccAddress{addr1, addr2, addr3}
	pubs := []cryptotypes.PubKey{pub1, pub2, pub3}

	msgs := make([]sdk.Msg, len(addrs))
	// set accounts and create msg for each address
	for i, addr := range addrs {
		acc := s.app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		require.NoError(acc.SetAccountNumber(uint64(i)))
		s.app.AccountKeeper.SetAccount(ctx, acc)
		msgs[i] = testdata.NewTestMsg(addr)
	}
	require.NoError(txBuilder.SetMsgs(msgs...))
	txBuilder.SetFeeAmount(testdata.NewTestFeeAmount())
	txBuilder.SetGasLimit(testdata.NewTestGasLimit())

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
	testTx, _, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
	require.NoError(err)

	_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), tx.Request{Tx: testTx})
	require.NoError(err)

	// Require that all accounts have pubkey set after middleware runs
	for i, addr := range addrs {
		pk, err := s.app.AccountKeeper.GetPubKey(ctx, addr)
		require.NoError(err, "Error on retrieving pubkey from account")
		require.True(pubs[i].Equals(pk),
			"Wrong Pubkey retrieved from AccountKeeper, idx=%d\nexpected=%s\n     got=%s", i, pubs[i], pk)
	}
}

func (s *MWTestSuite) TestConsumeSignatureVerificationGas() {
	params := types.DefaultParams()
	msg := []byte{1, 2, 3, 4}
	cdc := simapp.MakeTestEncodingConfig().Amino

	p := types.DefaultParams()
	skR1, _ := secp256r1.GenPrivKey()
	pkSet1, sigSet1 := generatePubKeysAndSignatures(5, msg, false)
	multisigKey1 := kmultisig.NewLegacyAminoPubKey(2, pkSet1)
	multisignature1 := multisig.NewMultisig(len(pkSet1))
	expectedCost1 := expectedGasCostByKeys(pkSet1)
	for i := 0; i < len(pkSet1); i++ {
		stdSig := legacytx.StdSignature{PubKey: pkSet1[i], Signature: sigSet1[i]}
		sigV2, err := legacytx.StdSignatureToSignatureV2(cdc, stdSig)
		s.Require().NoError(err)
		err = multisig.AddSignatureV2(multisignature1, sigV2, pkSet1)
		s.Require().NoError(err)
	}

	type args struct {
		meter  sdk.GasMeter
		sig    signing.SignatureData
		pubkey cryptotypes.PubKey
		params types.Params
	}
	tests := []struct {
		name        string
		args        args
		gasConsumed uint64
		shouldErr   bool
	}{
		{"PubKeyEd25519", args{sdk.NewInfiniteGasMeter(), nil, ed25519.GenPrivKey().PubKey(), params}, p.SigVerifyCostED25519, true},
		{"PubKeySecp256k1", args{sdk.NewInfiniteGasMeter(), nil, secp256k1.GenPrivKey().PubKey(), params}, p.SigVerifyCostSecp256k1, false},
		{"PubKeySecp256r1", args{sdk.NewInfiniteGasMeter(), nil, skR1.PubKey(), params}, p.SigVerifyCostSecp256r1(), false},
		{"Multisig", args{sdk.NewInfiniteGasMeter(), multisignature1, multisigKey1, params}, expectedCost1, false},
		{"unknown key", args{sdk.NewInfiniteGasMeter(), nil, nil, params}, 0, true},
	}
	for _, tt := range tests {
		sigV2 := signing.SignatureV2{
			PubKey:   tt.args.pubkey,
			Data:     tt.args.sig,
			Sequence: 0, // Arbitrary account sequence
		}
		err := middleware.DefaultSigVerificationGasConsumer(tt.args.meter, sigV2, tt.args.params)

		if tt.shouldErr {
			s.Require().NotNil(err)
		} else {
			s.Require().Nil(err)
			s.Require().Equal(tt.gasConsumed, tt.args.meter.GasConsumed(), fmt.Sprintf("%d != %d", tt.gasConsumed, tt.args.meter.GasConsumed()))
		}
	}
}

func (s *MWTestSuite) TestSigVerification() {
	ctx := s.SetupTest(true) // setup

	// make block height non-zero to ensure account numbers part of signBytes
	ctx = ctx.WithBlockHeight(1)
	txHandler := middleware.ComposeMiddlewares(
		noopTxHandler,
		middleware.SetPubKeyMiddleware(s.app.AccountKeeper),
		middleware.SigVerificationMiddleware(
			s.app.AccountKeeper,
			s.clientCtx.TxConfig.SignModeHandler(),
		),
	)

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()
	priv3, _, addr3 := testdata.KeyTestPubAddr()

	addrs := []sdk.AccAddress{addr1, addr2, addr3}

	msgs := make([]sdk.Msg, len(addrs))
	// set accounts and create msg for each address
	for i, addr := range addrs {
		acc := s.app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		s.Require().NoError(acc.SetAccountNumber(uint64(i)))
		s.app.AccountKeeper.SetAccount(ctx, acc)
		msgs[i] = testdata.NewTestMsg(addr)
	}

	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	type testCase struct {
		name      string
		privs     []cryptotypes.PrivKey
		accNums   []uint64
		accSeqs   []uint64
		recheck   bool
		shouldErr bool
	}
	testCases := []testCase{
		{"no signers", []cryptotypes.PrivKey{}, []uint64{}, []uint64{}, false, true},
		{"not enough signers", []cryptotypes.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{0, 0}, false, true},
		{"wrong order signers", []cryptotypes.PrivKey{priv3, priv2, priv1}, []uint64{2, 1, 0}, []uint64{0, 0, 0}, false, true},
		{"wrong accnums", []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{7, 8, 9}, []uint64{0, 0, 0}, false, true},
		{"wrong sequences", []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{3, 4, 5}, false, true},
		{"valid tx", []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}, false, false},
		{"no err on recheck", []cryptotypes.PrivKey{}, []uint64{}, []uint64{}, true, false},
	}
	for i, tc := range testCases {
		ctx = ctx.WithIsReCheckTx(tc.recheck)
		txBuilder := s.clientCtx.TxConfig.NewTxBuilder() // Create new txBuilder for each test

		s.Require().NoError(txBuilder.SetMsgs(msgs...))
		txBuilder.SetFeeAmount(feeAmount)
		txBuilder.SetGasLimit(gasLimit)

		testTx, _, err := s.createTestTx(txBuilder, tc.privs, tc.accNums, tc.accSeqs, ctx.ChainID())
		s.Require().NoError(err)

		if tc.recheck {
			_, _, err = txHandler.CheckTx(sdk.WrapSDKContext(ctx), tx.Request{Tx: testTx}, tx.RequestCheckTx{Type: abci.CheckTxType_Recheck})
		} else {
			_, _, err = txHandler.CheckTx(sdk.WrapSDKContext(ctx), tx.Request{Tx: testTx}, tx.RequestCheckTx{})
		}
		if tc.shouldErr {
			s.Require().NotNil(err, "TestCase %d: %s did not error as expected", i, tc.name)
		} else {
			s.Require().Nil(err, "TestCase %d: %s errored unexpectedly. Err: %v", i, tc.name, err)
		}
	}
}

// This test is exactly like the one above, but we set the codec explicitly to
// Amino.
// Once https://github.com/cosmos/cosmos-sdk/issues/6190 is in, we can remove
// this, since it'll be handled by the test matrix.
// In the meantime, we want to make double-sure amino compatibility works.
// ref: https://github.com/cosmos/cosmos-sdk/issues/7229
func (s *MWTestSuite) TestSigVerification_ExplicitAmino() {
	ctx := s.SetupTest(true)
	ctx = ctx.WithBlockHeight(1)

	// Set up TxConfig.
	aminoCdc := legacy.Cdc
	aminoCdc.RegisterInterface((*sdk.Msg)(nil), nil)
	aminoCdc.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)

	// We're using TestMsg amino encoding in some tests, so register it here.
	txConfig := legacytx.StdTxConfig{Cdc: aminoCdc}

	s.clientCtx = client.Context{}.
		WithTxConfig(txConfig)

	// make block height non-zero to ensure account numbers part of signBytes
	ctx = ctx.WithBlockHeight(1)

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()
	priv3, _, addr3 := testdata.KeyTestPubAddr()

	addrs := []sdk.AccAddress{addr1, addr2, addr3}

	msgs := make([]sdk.Msg, len(addrs))
	// set accounts and create msg for each address
	for i, addr := range addrs {
		acc := s.app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		s.Require().NoError(acc.SetAccountNumber(uint64(i)))
		s.app.AccountKeeper.SetAccount(ctx, acc)
		msgs[i] = testdata.NewTestMsg(addr)
	}

	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	txHandler := middleware.ComposeMiddlewares(
		noopTxHandler,
		middleware.SetPubKeyMiddleware(s.app.AccountKeeper),
		middleware.SigVerificationMiddleware(
			s.app.AccountKeeper,
			s.clientCtx.TxConfig.SignModeHandler(),
		),
	)

	type testCase struct {
		name      string
		privs     []cryptotypes.PrivKey
		accNums   []uint64
		accSeqs   []uint64
		recheck   bool
		shouldErr bool
	}

	testCases := []testCase{
		{"no signers", []cryptotypes.PrivKey{}, []uint64{}, []uint64{}, false, true},
		{"not enough signers", []cryptotypes.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{0, 0}, false, true},
		{"wrong order signers", []cryptotypes.PrivKey{priv3, priv2, priv1}, []uint64{2, 1, 0}, []uint64{0, 0, 0}, false, true},
		{"wrong accnums", []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{7, 8, 9}, []uint64{0, 0, 0}, false, true},
		{"wrong sequences", []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{3, 4, 5}, false, true},
		{"valid tx", []cryptotypes.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}, false, false},
		{"no err on recheck", []cryptotypes.PrivKey{}, []uint64{}, []uint64{}, true, false},
	}

	for i, tc := range testCases {
		ctx = ctx.WithIsReCheckTx(tc.recheck)
		txBuilder := s.clientCtx.TxConfig.NewTxBuilder() // Create new txBuilder for each test

		s.Require().NoError(txBuilder.SetMsgs(msgs...))
		txBuilder.SetFeeAmount(feeAmount)
		txBuilder.SetGasLimit(gasLimit)

		testTx, _, err := s.createTestTx(txBuilder, tc.privs, tc.accNums, tc.accSeqs, ctx.ChainID())
		s.Require().NoError(err)

		if tc.recheck {
			_, _, err = txHandler.CheckTx(sdk.WrapSDKContext(ctx), tx.Request{Tx: testTx}, tx.RequestCheckTx{Type: abci.CheckTxType_Recheck})
		} else {
			_, _, err = txHandler.CheckTx(sdk.WrapSDKContext(ctx), tx.Request{Tx: testTx}, tx.RequestCheckTx{})
		}
		if tc.shouldErr {
			s.Require().NotNil(err, "TestCase %d: %s did not error as expected", i, tc.name)
		} else {
			s.Require().Nil(err, "TestCase %d: %s errored unexpectedly. Err: %v", i, tc.name, err)
		}
	}
}

func (s *MWTestSuite) TestSigIntegration() {
	// generate private keys
	privs := []cryptotypes.PrivKey{
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
	}

	params := types.DefaultParams()
	initialSigCost := params.SigVerifyCostSecp256k1
	initialCost, err := s.runSigMiddlewares(params, false, privs...)
	s.Require().Nil(err)

	params.SigVerifyCostSecp256k1 *= 2
	doubleCost, err := s.runSigMiddlewares(params, false, privs...)
	s.Require().Nil(err)

	s.Require().Equal(initialSigCost*uint64(len(privs)), doubleCost-initialCost)
}

func (s *MWTestSuite) runSigMiddlewares(params types.Params, _ bool, privs ...cryptotypes.PrivKey) (sdk.Gas, error) {
	ctx := s.SetupTest(true) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	// Make block-height non-zero to include accNum in SignBytes
	ctx = ctx.WithBlockHeight(1)
	s.app.AccountKeeper.SetParams(ctx, params)

	msgs := make([]sdk.Msg, len(privs))
	accNums := make([]uint64, len(privs))
	accSeqs := make([]uint64, len(privs))
	// set accounts and create msg for each address
	for i, priv := range privs {
		addr := sdk.AccAddress(priv.PubKey().Address())
		acc := s.app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		s.Require().NoError(acc.SetAccountNumber(uint64(i)))
		s.app.AccountKeeper.SetAccount(ctx, acc)
		msgs[i] = testdata.NewTestMsg(addr)
		accNums[i] = uint64(i)
		accSeqs[i] = uint64(0)
	}
	s.Require().NoError(txBuilder.SetMsgs(msgs...))

	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)

	testTx, _, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
	s.Require().NoError(err)

	txHandler := middleware.ComposeMiddlewares(
		noopTxHandler,
		middleware.SetPubKeyMiddleware(s.app.AccountKeeper),
		middleware.SigGasConsumeMiddleware(s.app.AccountKeeper, middleware.DefaultSigVerificationGasConsumer),
		middleware.SigVerificationMiddleware(
			s.app.AccountKeeper,
			s.clientCtx.TxConfig.SignModeHandler(),
		),
	)

	// Determine gas consumption of txhandler with default params
	before := ctx.GasMeter().GasConsumed()
	_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), tx.Request{Tx: testTx})
	after := ctx.GasMeter().GasConsumed()

	return after - before, err
}

func (s *MWTestSuite) TestIncrementSequenceMiddleware() {
	ctx := s.SetupTest(true) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	priv, _, addr := testdata.KeyTestPubAddr()
	acc := s.app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	s.Require().NoError(acc.SetAccountNumber(uint64(50)))
	s.app.AccountKeeper.SetAccount(ctx, acc)

	msgs := []sdk.Msg{testdata.NewTestMsg(addr)}
	s.Require().NoError(txBuilder.SetMsgs(msgs...))
	privs := []cryptotypes.PrivKey{priv}
	accNums := []uint64{s.app.AccountKeeper.GetAccount(ctx, addr).GetAccountNumber()}
	accSeqs := []uint64{s.app.AccountKeeper.GetAccount(ctx, addr).GetSequence()}
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)

	testTx, _, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
	s.Require().NoError(err)

	txHandler := middleware.ComposeMiddlewares(
		noopTxHandler,
		middleware.IncrementSequenceMiddleware(s.app.AccountKeeper),
	)

	testCases := []struct {
		ctx         sdk.Context
		simulate    bool
		expectedSeq uint64
	}{
		{ctx.WithIsReCheckTx(true), false, 1},
		{ctx.WithIsCheckTx(true).WithIsReCheckTx(false), false, 2},
		{ctx.WithIsReCheckTx(true), false, 3},
		{ctx.WithIsReCheckTx(true), false, 4},
		{ctx.WithIsReCheckTx(true), true, 5},
	}

	for i, tc := range testCases {
		var err error
		if tc.simulate {
			_, err = txHandler.SimulateTx(sdk.WrapSDKContext(tc.ctx), tx.Request{Tx: testTx})
		} else {
			_, err = txHandler.DeliverTx(sdk.WrapSDKContext(tc.ctx), tx.Request{Tx: testTx})
		}

		s.Require().NoError(err, "unexpected error; tc #%d, %v", i, tc)
		s.Require().Equal(tc.expectedSeq, s.app.AccountKeeper.GetAccount(ctx, addr).GetSequence())
	}
}
