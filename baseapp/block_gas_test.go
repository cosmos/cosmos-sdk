package baseapp_test

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

var blockMaxGas = uint64(simapp.DefaultConsensusParams.Block.MaxGas)

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
		{"consume MaxUint64", math.MaxUint64, false, true},
		{"consume MaxGasWanted", txtypes.MaxGasWanted, false, true},
		{"consume block gas when paniced", 10, true, true},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var app *simapp.SimApp
			routerOpt := func(bapp *baseapp.BaseApp) {
				route := (&testdata.TestMsg{}).Route()
				bapp.Router().AddRoute(sdk.NewRoute(route, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
					_, ok := msg.(*testdata.TestMsg)
					if !ok {
						return &sdk.Result{}, fmt.Errorf("Wrong Msg type, expected %T, got %T", (*testdata.TestMsg)(nil), msg)
					}
					ctx.KVStore(app.GetKey(banktypes.ModuleName)).Set([]byte("ok"), []byte("ok"))
					ctx.GasMeter().ConsumeGas(tc.gasToConsume, "TestMsg")
					if tc.panicTx {
						panic("panic in tx execution")
					}
					return &sdk.Result{}, nil
				}))
			}
			encCfg := simapp.MakeTestEncodingConfig()
			encCfg.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)
			encCfg.InterfaceRegistry.RegisterImplementations((*sdk.Msg)(nil),
				&testdata.TestMsg{},
			)
			app = simapp.NewSimApp(log.NewNopLogger(), dbm.NewMemDB(), nil, true, map[int64]bool{}, "", 0, encCfg, simapp.EmptyAppOptions{}, routerOpt)
			genState := simapp.NewDefaultGenesisState(encCfg.Marshaler)
			stateBytes, err := json.MarshalIndent(genState, "", " ")
			require.NoError(t, err)
			app.InitChain(abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: simapp.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			})

			ctx := app.NewContext(false, tmproto.Header{})

			// tx fee
			feeCoin := sdk.NewCoin("atom", sdk.NewInt(150))
			feeAmount := sdk.NewCoins(feeCoin)

			// test account and fund
			priv1, _, addr1 := testdata.KeyTestPubAddr()
			err = app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, feeAmount)
			require.NoError(t, err)
			err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr1, feeAmount)
			require.NoError(t, err)
			require.Equal(t, feeCoin.Amount, app.BankKeeper.GetBalance(ctx, addr1, feeCoin.Denom).Amount)
			seq, _ := app.AccountKeeper.GetSequence(ctx, addr1)
			require.Equal(t, uint64(0), seq)

			// msg and signatures
			msg := testdata.NewTestMsg(addr1)

			txBuilder := encCfg.TxConfig.NewTxBuilder()
			require.NoError(t, txBuilder.SetMsgs(msg))
			txBuilder.SetFeeAmount(feeAmount)
			txBuilder.SetGasLimit(txtypes.MaxGasWanted) // tx validation checks that gasLimit can't be bigger than this

			privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{6}, []uint64{0}
			_, txBytes, err := createTestTx(encCfg.TxConfig, txBuilder, privs, accNums, accSeqs, ctx.ChainID())
			require.NoError(t, err)

			app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 1}})
			rsp := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})

			// check result
			ctx = app.GetContextForDeliverTx(txBytes)
			okValue := ctx.KVStore(app.GetKey(banktypes.ModuleName)).Get([]byte("ok"))

			if tc.expErr {
				if tc.panicTx {
					require.Equal(t, sdkerrors.ErrPanic.ABCICode(), rsp.Code)
				} else {
					require.Equal(t, sdkerrors.ErrOutOfGas.ABCICode(), rsp.Code)
				}
				require.Empty(t, okValue)
			} else {
				require.Equal(t, uint32(0), rsp.Code)
				require.Equal(t, []byte("ok"), okValue)
			}
			// check block gas is always consumed
			baseGas := uint64(59142) // baseGas is the gas consumed before tx msg
			expGasConsumed := addUint64Saturating(tc.gasToConsume, baseGas)
			if expGasConsumed > txtypes.MaxGasWanted {
				// capped by gasLimit
				expGasConsumed = txtypes.MaxGasWanted
			}
			require.Equal(t, expGasConsumed, ctx.BlockGasMeter().GasConsumed())
			// tx fee is always deducted
			require.Equal(t, int64(0), app.BankKeeper.GetBalance(ctx, addr1, feeCoin.Denom).Amount.Int64())
			// sender's sequence is always increased
			seq, err = app.AccountKeeper.GetSequence(ctx, addr1)
			require.NoError(t, err)
			require.Equal(t, uint64(1), seq)
		})
	}
}

func createTestTx(txConfig client.TxConfig, txBuilder client.TxBuilder, privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (xauthsigning.Tx, []byte, error) {
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  txConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err := txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sigV2, err := tx.SignWithPrivKey(
			txConfig.SignModeHandler().DefaultMode(), signerData,
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
