package baseapp_test

import (
	"context"
	"math"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	store "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/app"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
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

		sdkAppConfig := app.DefaultSDKAppConfig("test", simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))
		bApp := app.NewSDKApp(log.NewTestLogger(t), dbm.NewMemDB(), nil, sdkAppConfig)

		bApp.LoadModules()
		require.NoError(t, bApp.LoadLatestVersion())

		t.Run(tc.name, func(t *testing.T) {
			baseapptestutil.RegisterInterfaces(bApp.InterfaceRegistry())
			baseapptestutil.RegisterKeyValueServer(bApp.MsgServiceRouter(), BlockGasImpl{
				panicTx:      tc.panicTx,
				gasToConsume: tc.gasToConsume,
				key:          bApp.UnsafeFindStoreKey(banktypes.ModuleName),
			})

			genState := genesisStateWithSingleValidator(t, bApp.AppCodec(), bApp)
			stateBytes, err := cmtjson.MarshalIndent(genState, "", " ")
			require.NoError(t, err)
			_, err = bApp.InitChain(&abci.RequestInitChain{
				ChainId:         bApp.ChainID(),
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: simtestutil.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			})

			require.NoError(t, err)
			ctx := bApp.NewContext(false).WithChainID(bApp.ChainID())

			// tx fee
			feeCoin := sdk.NewCoin("atom", sdkmath.NewInt(150))
			feeAmount := sdk.NewCoins(feeCoin)

			// test account and fund
			priv1, _, addr1 := testdata.KeyTestPubAddr()
			err = bApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, feeAmount)
			require.NoError(t, err)
			err = bApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr1, feeAmount)
			require.NoError(t, err)
			require.Equal(t, feeCoin.Amount, bApp.BankKeeper.GetBalance(ctx, addr1, feeCoin.Denom).Amount)
			seq := bApp.AccountKeeper.GetAccount(ctx, addr1).GetSequence()
			require.Equal(t, uint64(0), seq)

			// msg and signatures
			msg := &baseapptestutil.MsgKeyValue{
				Key:    []byte("ok"),
				Value:  []byte("ok"),
				Signer: addr1.String(),
			}

			txBuilder := bApp.TxConfig().NewTxBuilder()

			require.NoError(t, txBuilder.SetMsgs(msg))
			txBuilder.SetFeeAmount(feeAmount)
			txBuilder.SetGasLimit(uint64(simtestutil.DefaultConsensusParams.Block.MaxGas))

			senderAccountNumber := bApp.AccountKeeper.GetAccount(ctx, addr1).GetAccountNumber()
			privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{senderAccountNumber}, []uint64{0}
			_, txBytes, err := createTestTx(bApp.TxConfig(), txBuilder, privs, accNums, accSeqs, ctx.ChainID())
			require.NoError(t, err)

			rsp, err := bApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Txs: [][]byte{txBytes}})
			require.NoError(t, err)

			// check result
			ctx = bApp.GetContextForFinalizeBlock(txBytes)
			okValue := ctx.KVStore(bApp.UnsafeFindStoreKey(banktypes.ModuleName)).Get([]byte("ok"))

			if tc.expErr {
				if tc.panicTx {
					require.Equal(t, sdkerrors.ErrPanic.ABCICode(), rsp.TxResults[0].Code, rsp.TxResults[0].Log)
				} else {
					require.Equal(t, sdkerrors.ErrOutOfGas.ABCICode(), rsp.TxResults[0].Code, rsp.TxResults[0].Log)
				}
				require.Empty(t, okValue)
			} else {
				require.Equal(t, uint32(0), rsp.TxResults[0].Code, rsp.TxResults[0].Log)
				require.Equal(t, []byte("ok"), okValue)
			}
			// check block gas is always consumed
			baseGas := uint64(57504) // baseGas is the gas consumed before tx msg
			expGasConsumed := min(addUint64Saturating(tc.gasToConsume, baseGas), uint64(simtestutil.DefaultConsensusParams.Block.MaxGas))
			require.Equal(t, int(expGasConsumed), int(ctx.BlockGasMeter().GasConsumed()))
			// tx fee is always deducted
			require.Equal(t, int64(0), bApp.BankKeeper.GetBalance(ctx, addr1, feeCoin.Denom).Amount.Int64())
			// sender's sequence is always increased
			seq = bApp.AccountKeeper.GetAccount(ctx, addr1).GetSequence()
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
