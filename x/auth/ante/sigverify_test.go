package ante_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestSetPubKey(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)

	// keys and addresses
	priv1, pub1, addr1 := types.KeyTestPubAddr()
	priv2, pub2, addr2 := types.KeyTestPubAddr()
	priv3, pub3, addr3 := types.KeyTestPubAddr()

	addrs := []sdk.AccAddress{addr1, addr2, addr3}
	pubs := []crypto.PubKey{pub1, pub2, pub3}

	msgs := make([]sdk.Msg, len(addrs))
	// set accounts and create msg for each address
	for i, addr := range addrs {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		require.NoError(t, acc.SetAccountNumber(uint64(i)))
		app.AccountKeeper.SetAccount(ctx, acc)
		msgs[i] = types.NewTestMsg(addr)
	}

	fee := types.NewTestStdFee()

	privs, accNums, seqs := []crypto.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
	tx := types.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

	spkd := ante.NewSetPubKeyDecorator(app.AccountKeeper)
	antehandler := sdk.ChainAnteDecorators(spkd)

	ctx, err := antehandler(ctx, tx, false)
	require.Nil(t, err)

	// Require that all accounts have pubkey set after Decorator runs
	for i, addr := range addrs {
		pk, err := app.AccountKeeper.GetPubKey(ctx, addr)
		require.Nil(t, err, "Error on retrieving pubkey from account")
		require.Equal(t, pubs[i], pk, "Pubkey retrieved from account is unexpected")
	}
}

func TestConsumeSignatureVerificationGas(t *testing.T) {
	params := types.DefaultParams()
	msg := []byte{1, 2, 3, 4}

	pkSet1, sigSet1 := generatePubKeysAndSignatures(5, msg, false)
	multisigKey1 := multisig.NewPubKeyMultisigThreshold(2, pkSet1)
	multisignature1 := multisig.NewMultisig(len(pkSet1))
	expectedCost1 := expectedGasCostByKeys(pkSet1)
	for i := 0; i < len(pkSet1); i++ {
		multisignature1.AddSignatureFromPubKey(sigSet1[i], pkSet1[i], pkSet1)
	}

	type args struct {
		meter  sdk.GasMeter
		sig    []byte
		pubkey crypto.PubKey
		params types.Params
	}
	tests := []struct {
		name        string
		args        args
		gasConsumed uint64
		shouldErr   bool
	}{
		{"PubKeyEd25519", args{sdk.NewInfiniteGasMeter(), nil, ed25519.GenPrivKey().PubKey(), params}, types.DefaultSigVerifyCostED25519, true},
		{"PubKeySecp256k1", args{sdk.NewInfiniteGasMeter(), nil, secp256k1.GenPrivKey().PubKey(), params}, types.DefaultSigVerifyCostSecp256k1, false},
		{"Multisig", args{sdk.NewInfiniteGasMeter(), multisignature1.Marshal(), multisigKey1, params}, expectedCost1, false},
		{"unknown key", args{sdk.NewInfiniteGasMeter(), nil, nil, params}, 0, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := ante.DefaultSigVerificationGasConsumer(tt.args.meter, tt.args.sig, tt.args.pubkey, tt.args.params)

			if tt.shouldErr {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
				require.Equal(t, tt.gasConsumed, tt.args.meter.GasConsumed(), fmt.Sprintf("%d != %d", tt.gasConsumed, tt.args.meter.GasConsumed()))
			}
		})
	}
}

func TestSigVerification(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)
	// make block height non-zero to ensure account numbers part of signBytes
	ctx = ctx.WithBlockHeight(1)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()
	priv3, _, addr3 := types.KeyTestPubAddr()

	addrs := []sdk.AccAddress{addr1, addr2, addr3}

	msgs := make([]sdk.Msg, len(addrs))
	// set accounts and create msg for each address
	for i, addr := range addrs {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		require.NoError(t, acc.SetAccountNumber(uint64(i)))
		app.AccountKeeper.SetAccount(ctx, acc)
		msgs[i] = types.NewTestMsg(addr)
	}

	fee := types.NewTestStdFee()

	spkd := ante.NewSetPubKeyDecorator(app.AccountKeeper)
	svd := ante.NewSigVerificationDecorator(app.AccountKeeper)
	antehandler := sdk.ChainAnteDecorators(spkd, svd)

	type testCase struct {
		name      string
		privs     []crypto.PrivKey
		accNums   []uint64
		seqs      []uint64
		recheck   bool
		shouldErr bool
	}
	testCases := []testCase{
		{"no signers", []crypto.PrivKey{}, []uint64{}, []uint64{}, false, true},
		{"not enough signers", []crypto.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{0, 0}, false, true},
		{"wrong order signers", []crypto.PrivKey{priv3, priv2, priv1}, []uint64{2, 1, 0}, []uint64{0, 0, 0}, false, true},
		{"wrong accnums", []crypto.PrivKey{priv1, priv2, priv3}, []uint64{7, 8, 9}, []uint64{0, 0, 0}, false, true},
		{"wrong sequences", []crypto.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{3, 4, 5}, false, true},
		{"valid tx", []crypto.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}, false, false},
		{"no err on recheck", []crypto.PrivKey{}, []uint64{}, []uint64{}, true, false},
	}
	for i, tc := range testCases {
		ctx = ctx.WithIsReCheckTx(tc.recheck)

		tx := types.NewTestTx(ctx, msgs, tc.privs, tc.accNums, tc.seqs, fee)

		_, err := antehandler(ctx, tx, false)
		if tc.shouldErr {
			require.NotNil(t, err, "TestCase %d: %s did not error as expected", i, tc.name)
		} else {
			require.Nil(t, err, "TestCase %d: %s errored unexpectedly. Err: %v", i, tc.name, err)
		}
	}
}

func TestSigIntegration(t *testing.T) {
	// generate private keys
	privs := []crypto.PrivKey{secp256k1.GenPrivKey(), secp256k1.GenPrivKey(), secp256k1.GenPrivKey()}

	params := types.DefaultParams()
	initialSigCost := params.SigVerifyCostSecp256k1
	initialCost, err := runSigDecorators(t, params, false, privs...)
	require.Nil(t, err)

	params.SigVerifyCostSecp256k1 *= 2
	doubleCost, err := runSigDecorators(t, params, false, privs...)
	require.Nil(t, err)

	require.Equal(t, initialSigCost*uint64(len(privs)), doubleCost-initialCost)
}

func runSigDecorators(t *testing.T, params types.Params, multisig bool, privs ...crypto.PrivKey) (sdk.Gas, error) {
	// setup
	app, ctx := createTestApp(true)
	// Make block-height non-zero to include accNum in SignBytes
	ctx = ctx.WithBlockHeight(1)
	app.AccountKeeper.SetParams(ctx, params)

	msgs := make([]sdk.Msg, len(privs))
	accNums := make([]uint64, len(privs))
	seqs := make([]uint64, len(privs))
	// set accounts and create msg for each address
	for i, priv := range privs {
		addr := sdk.AccAddress(priv.PubKey().Address())
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		require.NoError(t, acc.SetAccountNumber(uint64(i)))
		app.AccountKeeper.SetAccount(ctx, acc)
		msgs[i] = types.NewTestMsg(addr)
		accNums[i] = uint64(i)
		seqs[i] = uint64(0)
	}

	fee := types.NewTestStdFee()

	tx := types.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

	spkd := ante.NewSetPubKeyDecorator(app.AccountKeeper)
	svgc := ante.NewSigGasConsumeDecorator(app.AccountKeeper, ante.DefaultSigVerificationGasConsumer)
	svd := ante.NewSigVerificationDecorator(app.AccountKeeper)
	antehandler := sdk.ChainAnteDecorators(spkd, svgc, svd)

	// Determine gas consumption of antehandler with default params
	before := ctx.GasMeter().GasConsumed()
	ctx, err := antehandler(ctx, tx, false)
	after := ctx.GasMeter().GasConsumed()

	return after - before, err
}

func TestIncrementSequenceDecorator(t *testing.T) {
	app, ctx := createTestApp(true)

	priv, _, addr := types.KeyTestPubAddr()
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	require.NoError(t, acc.SetAccountNumber(uint64(50)))
	app.AccountKeeper.SetAccount(ctx, acc)

	msgs := []sdk.Msg{types.NewTestMsg(addr)}
	privKeys := []crypto.PrivKey{priv}
	accNums := []uint64{app.AccountKeeper.GetAccount(ctx, addr).GetAccountNumber()}
	accSeqs := []uint64{app.AccountKeeper.GetAccount(ctx, addr).GetSequence()}
	fee := types.NewTestStdFee()
	tx := types.NewTestTx(ctx, msgs, privKeys, accNums, accSeqs, fee)

	isd := ante.NewIncrementSequenceDecorator(app.AccountKeeper)
	antehandler := sdk.ChainAnteDecorators(isd)

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
		_, err := antehandler(tc.ctx, tx, tc.simulate)
		require.NoError(t, err, "unexpected error; tc #%d, %v", i, tc)
		require.Equal(t, tc.expectedSeq, app.AccountKeeper.GetAccount(ctx, addr).GetSequence())
	}
}
