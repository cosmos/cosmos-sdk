package ante_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"
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

	var msgs []sdk.Msg
	// set accounts and create msg for each address
	for i, addr := range addrs {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		require.NoError(t, acc.SetAccountNumber(uint64(i)))
		app.AccountKeeper.SetAccount(ctx, acc)
		msgs = append(msgs, types.NewTestMsg(addr))
	}

	fee := types.NewTestStdFee()

	privs, accNums, seqs := []crypto.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
	tx := types.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

	spkd := ante.NewSetPubKeyDecorator(app.AccountKeeper)
	antehandler := sdk.ChainDecorators(spkd)

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

func TestSigGasConsumerSecp56k1(t *testing.T) {
	// generate private keys
	privs := []crypto.PrivKey{secp256k1.GenPrivKey(), secp256k1.GenPrivKey(), secp256k1.GenPrivKey()}

	params := types.DefaultParams()
	initialSigCost := params.SigVerifyCostSecp256k1
	initialCost := benchmarkGasSigDecorator(t, params, false, privs...)

	params.SigVerifyCostSecp256k1 = params.SigVerifyCostSecp256k1 * 2
	doubleCost := benchmarkGasSigDecorator(t, params, false, privs...)

	require.Equal(t, initialSigCost*uint64(len(privs)), doubleCost-initialCost)
}

func TestSigVerification(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()
	priv3, _, addr3 := types.KeyTestPubAddr()

	addrs := []sdk.AccAddress{addr1, addr2, addr3}

	var msgs []sdk.Msg
	// set accounts and create msg for each address
	for i, addr := range addrs {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		require.NoError(t, acc.SetAccountNumber(uint64(i)))
		app.AccountKeeper.SetAccount(ctx, acc)
		msgs = append(msgs, types.NewTestMsg(addr))
	}

	fee := types.NewTestStdFee()

	privs, accNums, seqs := []crypto.PrivKey{priv3, priv2, priv1}, []uint64{2, 1, 0}, []uint64{0, 0, 0}
	tx := types.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

	spkd := ante.NewSetPubKeyDecorator(app.AccountKeeper)
	svd := ante.NewSigVerificationDecorator(app.AccountKeeper)
	antehandler := sdk.ChainDecorators(spkd, svd)

	ctx, err := antehandler(ctx, tx, false)
	require.NotNil(t, err)
}

func benchmarkGasSigDecorator(t *testing.T, params types.Params, multisig bool, privs ...crypto.PrivKey) sdk.Gas {
	// setup
	app, ctx := createTestApp(true)
	app.AccountKeeper.SetParams(ctx, params)

	var msgs []sdk.Msg
	var accNums []uint64
	var seqs []uint64
	// set accounts and create msg for each address
	for i, priv := range privs {
		addr := sdk.AccAddress(priv.PubKey().Address())
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		require.NoError(t, acc.SetAccountNumber(uint64(i)))
		app.AccountKeeper.SetAccount(ctx, acc)
		msgs = append(msgs, types.NewTestMsg(addr))
		accNums = append(accNums, uint64(i))
		seqs = append(seqs, uint64(0))
	}

	fee := types.NewTestStdFee()

	tx := types.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

	spkd := ante.NewSetPubKeyDecorator(app.AccountKeeper)
	svgc := ante.NewSigGasConsumeDecorator(app.AccountKeeper, ante.DefaultSigVerificationGasConsumer)
	antehandler := sdk.ChainDecorators(spkd, svgc)

	// Determine gas consumption of antehandler with default params
	before := ctx.GasMeter().GasConsumed()
	ctx, err := antehandler(ctx, tx, false)
	require.Nil(t, err)
	after := ctx.GasMeter().GasConsumed()

	return sdk.Gas(after - before)
}
