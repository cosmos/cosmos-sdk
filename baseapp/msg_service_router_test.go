package baseapp_test

import (
	"context"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	txsigning "github.com/cosmos/cosmos-sdk/x/tx/signing"
)

// newTestBaseApp creates a minimal BaseApp for router unit tests (no modules, no InitChain).
func newTestBaseApp(t *testing.T, logger log.Logger) (*baseapp.BaseApp, codectypes.InterfaceRegistry) {
	t.Helper()

	interfaceRegistry, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: gogoproto.HybridResolver,
		SigningOptions: txsigning.Options{
			AddressCodec:          address.Bech32Codec{Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix()},
			ValidatorAddressCodec: address.Bech32Codec{Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix()},
		},
	})
	require.NoError(t, err)
	std.RegisterInterfaces(interfaceRegistry)

	cdc := codec.NewProtoCodec(interfaceRegistry)
	app := baseapp.NewBaseApp("test", logger, dbm.NewMemDB(), nil)
	app.SetInterfaceRegistry(interfaceRegistry)
	_ = cdc

	return app, interfaceRegistry
}

func TestRegisterMsgService(t *testing.T) {
	app, registry := newTestBaseApp(t, log.NewTestLogger(t))

	require.Panics(t, func() {
		testdata.RegisterMsgServer(
			app.MsgServiceRouter(),
			testdata.MsgServerImpl{},
		)
	})

	// Register testdata Msg services, and rerun `RegisterMsgService`.
	testdata.RegisterInterfaces(registry)

	require.NotPanics(t, func() {
		testdata.RegisterMsgServer(
			app.MsgServiceRouter(),
			testdata.MsgServerImpl{},
		)
	})
}

func TestRegisterMsgServiceTwice(t *testing.T) {
	app, registry := newTestBaseApp(t, log.NewTestLogger(t))
	testdata.RegisterInterfaces(registry)

	// First time registering service shouldn't panic.
	require.NotPanics(t, func() {
		testdata.RegisterMsgServer(
			app.MsgServiceRouter(),
			testdata.MsgServerImpl{},
		)
	})

	// Second time should panic.
	require.Panics(t, func() {
		testdata.RegisterMsgServer(
			app.MsgServiceRouter(),
			testdata.MsgServerImpl{},
		)
	})
}

func TestHybridHandlerByMsgName(t *testing.T) {
	app, registry := newTestBaseApp(t, log.NewTestLogger(t))
	testdata.RegisterInterfaces(registry)

	testdata.RegisterMsgServer(
		app.MsgServiceRouter(),
		testdata.MsgServerImpl{},
	)

	handler := app.MsgServiceRouter().HybridHandlerByMsgName("testpb.MsgCreateDog")

	require.NotNil(t, handler)
	require.NoError(t, app.Init())
	ctx := app.NewContext(true)
	resp := new(testdata.MsgCreateDogResponse)
	err := handler(ctx, &testdata.MsgCreateDog{
		Dog:   &testdata.Dog{Name: "Spot"},
		Owner: "me",
	}, resp)
	require.NoError(t, err)
	require.Equal(t, resp.Name, "Spot")
}

func TestMsgService(t *testing.T) {
	priv, _, _ := testdata.KeyTestPubAddr()

	app, interfaceRegistry := newTestBaseApp(t, log.NewNopLogger())
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// patch in TxConfig instead of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	// set the TxDecoder in the BaseApp for minimal tx simulations
	app.SetTxDecoder(txConfig.TxDecoder())

	defaultSignMode, err := authsigning.APISignModeToInternal(txConfig.SignModeHandler().DefaultMode())
	require.NoError(t, err)

	testdata.RegisterInterfaces(interfaceRegistry)
	testdata.RegisterMsgServer(
		app.MsgServiceRouter(),
		testdata.MsgServerImpl{},
	)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)

	_, _, addr := testdata.KeyTestPubAddr()
	msg := testdata.MsgCreateDog{
		Dog:   &testdata.Dog{Name: "Spot"},
		Owner: addr.String(),
	}

	txBuilder := txConfig.NewTxBuilder()
	txBuilder.SetFeeAmount(testdata.NewTestFeeAmount())
	txBuilder.SetGasLimit(testdata.NewTestGasLimit())
	err = txBuilder.SetMsgs(&msg)
	require.NoError(t, err)

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signtypes.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signtypes.SingleSignatureData{
			SignMode:  defaultSignMode,
			Signature: nil,
		},
		Sequence: 0,
	}

	err = txBuilder.SetSignatures(sigV2)
	require.NoError(t, err)

	// Second round: all signer infos are set, so each signer can sign.
	signerData := authsigning.SignerData{
		ChainID:       "test",
		AccountNumber: 0,
		Sequence:      0,
		PubKey:        priv.PubKey(),
	}
	sigV2, err = tx.SignWithPrivKey(
		context.TODO(), defaultSignMode, signerData,
		txBuilder, priv, txConfig, 0)
	require.NoError(t, err)
	err = txBuilder.SetSignatures(sigV2)
	require.NoError(t, err)

	// Send the tx to the app
	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	require.NoError(t, err)
	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Txs: [][]byte{txBytes}})
	require.NoError(t, err)
	require.Equal(t, abci.CodeTypeOK, res.TxResults[0].Code, "res=%+v", res)
}
