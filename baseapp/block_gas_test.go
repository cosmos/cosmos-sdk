package baseapp_test

import (
	"context"
	"math"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"

	sdkapp "github.com/cosmos/cosmos-sdk/app"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	store "github.com/cosmos/cosmos-sdk/store/v2/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var blockMaxGas = uint64(simtestutil.DefaultConsensusParams.Block.MaxGas)

type BlockGasImpl struct {
	panicTx      bool
	gasToConsume uint64
	key          store.StoreKey
}

func (m BlockGasImpl) Set(ctx context.Context, msg *baseapptestutil.MsgKeyValue) (*baseapptestutil.MsgCreateKeyValueResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.KVStore(m.key).Set(msg.Key, msg.Value)
	sdkCtx.GasMeter().ConsumeGas(m.gasToConsume, "TestMsg")
	if m.panicTx {
		panic("panic in tx execution")
	}
	return &baseapptestutil.MsgCreateKeyValueResponse{}, nil
}

// newBlockGasApp creates a fresh SDKApp per test iteration (before InitChain) so that
// we can register custom msg handlers and supply genesis before the app starts.
func newBlockGasApp(t *testing.T) (
	bapp *sdkapp.SDKApp,
	bankKeeper bankkeeper.Keeper,
	accountKeeper authkeeper.AccountKeeper,
	txConfig client.TxConfig,
	cdc codec.Codec,
	registry codectypes.InterfaceRegistry,
) {
	t.Helper()

	opts := simtestutil.AppOptionsMap{
		flags.FlagHome:    t.TempDir(),
		flags.FlagChainID: "test-chain",
	}
	cfg := sdkapp.DefaultSDKAppConfig("app", opts)
	app := sdkapp.NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)
	app.LoadModules()
	if err := app.LoadLatestVersion(); err != nil {
		t.Fatalf("failed to load latest version: %v", err)
	}

	app.SetDisableBlockGasMeter(false)

	return app,
		app.BankKeeper,
		app.AccountKeeper,
		app.TxConfig(),
		app.AppCodec(),
		app.InterfaceRegistry()
}

func TestBaseApp_BlockGas(t *testing.T) {
	testcases := []struct {
		name         string
		gasToConsume uint64 // gas to consume in the msg execution
		panicTx      bool   // panic explicitly in tx execution
		expErr       bool
	}{
		{"less than block gas meter", 10, false, false},
		{"more than block gas meter", blockMaxGas, false, true},
		{"more than block gas meter", uint64(float64(blockMaxGas) * 1.2), false, true},
		{"consume MaxUint64", math.MaxUint64, true, true},
		{"consume MaxGasWanted", txtypes.MaxGasWanted, false, true},
		{"consume block gas when panicked", 10, true, true},
	}

	for _, tc := range testcases {
		priv1, _, addr1 := testdata.KeyTestPubAddr()
		feeCoin := sdk.NewCoin("atom", sdkmath.NewInt(150))
		feeAmount := sdk.NewCoins(feeCoin)

		bapp, bankKeeper, accountKeeper, txConfig, cdc, interfaceRegistry := newBlockGasApp(t)

		t.Run(tc.name, func(t *testing.T) {
			// Find bank store key to use as a scratch KV store in the test handler.
			var bankStoreKey store.StoreKey
			for _, k := range bapp.GetStoreKeys() {
				if k.Name() == banktypes.ModuleName {
					bankStoreKey = k
					break
				}
			}
			require.NotNil(t, bankStoreKey)

			baseapptestutil.RegisterInterfaces(interfaceRegistry)
			baseapptestutil.RegisterKeyValueServer(bapp.MsgServiceRouter(), BlockGasImpl{
				panicTx:      tc.panicTx,
				gasToConsume: tc.gasToConsume,
				key:          bankStoreKey,
			})

			// Include addr1 in genesis so its account number is committed before FinalizeBlock.
			genState := GenesisStateWithValidatorAndFeeAccount(t, cdc, bapp, priv1.PubKey(), feeAmount)
			stateBytes, err := cmtjson.MarshalIndent(genState, "", " ")
			require.NoError(t, err)
			_, err = bapp.InitChain(&abci.RequestInitChain{
				ChainId:         "test-chain",
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: simtestutil.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			})
			require.NoError(t, err)

			ctx := bapp.NewContext(false)

			acc1 := accountKeeper.GetAccount(ctx, addr1)
			require.NotNil(t, acc1)
			seq := acc1.GetSequence()
			require.Equal(t, uint64(0), seq)

			// msg and signatures
			msg := &baseapptestutil.MsgKeyValue{
				Key:    []byte("ok"),
				Value:  []byte("ok"),
				Signer: addr1.String(),
			}

			txBuilder := txConfig.NewTxBuilder()
			require.NoError(t, txBuilder.SetMsgs(msg))
			txBuilder.SetFeeAmount(feeAmount)
			txBuilder.SetGasLimit(uint64(simtestutil.DefaultConsensusParams.Block.MaxGas))

			senderAccountNumber := accountKeeper.GetAccount(ctx, addr1).GetAccountNumber()
			privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{senderAccountNumber}, []uint64{0}
			_, txBytes, err := createTestTx(txConfig, txBuilder, privs, accNums, accSeqs, bapp.ChainID())
			require.NoError(t, err)

			rsp, err := bapp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Txs: [][]byte{txBytes}})
			require.NoError(t, err)

			// check result
			ctx = bapp.GetContextForFinalizeBlock(txBytes)
			okValue := ctx.KVStore(bankStoreKey).Get([]byte("ok"))

			if tc.expErr {
				if tc.panicTx {
					require.Equal(t, sdkerrors.ErrPanic.ABCICode(), rsp.TxResults[0].Code)
				} else {
					require.Equal(t, sdkerrors.ErrOutOfGas.ABCICode(), rsp.TxResults[0].Code)
				}
				require.Empty(t, okValue)
			} else {
				require.Equal(t, uint32(0), rsp.TxResults[0].Code)
				require.Equal(t, []byte("ok"), okValue)
			}
			// check block gas is always consumed
			baseGas := uint64(44189) // baseGas is the gas consumed by ante handler before msg execution
			expGasConsumed := min(addUint64Saturating(tc.gasToConsume, baseGas), uint64(simtestutil.DefaultConsensusParams.Block.MaxGas))
			require.Equal(t, int(expGasConsumed), int(ctx.BlockGasMeter().GasConsumed()))
			// tx fee is always deducted
			require.Equal(t, int64(0), bankKeeper.GetBalance(ctx, addr1, feeCoin.Denom).Amount.Int64())
			// sender's sequence is always increased
			seq = accountKeeper.GetAccount(ctx, addr1).GetSequence()
			require.NoError(t, err)
			require.Equal(t, uint64(1), seq)
		})
	}
}

func createTestTx(txConfig client.TxConfig, txBuilder client.TxBuilder, privs []cryptotypes.PrivKey, accNums, accSeqs []uint64, chainID string) (xauthsigning.Tx, []byte, error) {
	defaultSignMode, err := xauthsigning.APISignModeToInternal(txConfig.SignModeHandler().DefaultMode())
	if err != nil {
		return nil, nil, err
	}
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  defaultSignMode,
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			Address:       sdk.AccAddress(priv.PubKey().Bytes()).String(),
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
			PubKey:        priv.PubKey(),
		}
		sigV2, err := tx.SignWithPrivKey(
			context.TODO(), defaultSignMode, signerData,
			txBuilder, priv, txConfig, accSeqs[i])
		if err != nil {
			return nil, nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, nil, err
	}

	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, nil, err
	}

	return txBuilder.GetTx(), txBytes, nil
}

func addUint64Saturating(a, b uint64) uint64 {
	if math.MaxUint64-a < b {
		return math.MaxUint64
	}

	return a + b
}
